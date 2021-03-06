/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package unstructured

import (
	gojson "encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
)

// NestedFieldCopy returns a deep copy of the value of a nested field.
// false is returned if the value is missing.
// nil, true is returned for a nil field.
func NestedFieldCopy(obj map[string]interface{}, fields ...string) (interface{}, bool) {
	val, ok := nestedFieldNoCopy(obj, fields...)
	if !ok {
		return nil, false
	}
	return runtime.DeepCopyJSONValue(val), true
}

func nestedFieldNoCopy(obj map[string]interface{}, fields ...string) (interface{}, bool) {
	var val interface{} = obj
	for _, field := range fields {
		if m, ok := val.(map[string]interface{}); ok {
			val, ok = m[field]
			if !ok {
				return nil, false
			}
		} else {
			// Expected map[string]interface{}, got something else
			return nil, false
		}
	}
	return val, true
}

// NestedString returns the string value of a nested field.
// Returns false if value is not found or is not a string.
func NestedString(obj map[string]interface{}, fields ...string) (string, bool) {
	val, ok := nestedFieldNoCopy(obj, fields...)
	if !ok {
		return "", false
	}
	s, ok := val.(string)
	return s, ok
}

// NestedBool returns the bool value of a nested field.
// Returns false if value is not found or is not a bool.
func NestedBool(obj map[string]interface{}, fields ...string) (bool, bool) {
	val, ok := nestedFieldNoCopy(obj, fields...)
	if !ok {
		return false, false
	}
	b, ok := val.(bool)
	return b, ok
}

// NestedFloat64 returns the bool value of a nested field.
// Returns false if value is not found or is not a float64.
func NestedFloat64(obj map[string]interface{}, fields ...string) (float64, bool) {
	val, ok := nestedFieldNoCopy(obj, fields...)
	if !ok {
		return 0, false
	}
	f, ok := val.(float64)
	return f, ok
}

// NestedInt64 returns the int64 value of a nested field.
// Returns false if value is not found or is not an int64.
func NestedInt64(obj map[string]interface{}, fields ...string) (int64, bool) {
	val, ok := nestedFieldNoCopy(obj, fields...)
	if !ok {
		return 0, false
	}
	i, ok := val.(int64)
	return i, ok
}

// NestedStringSlice returns a copy of []string value of a nested field.
// Returns false if value is not found, is not a []interface{} or contains non-string items in the slice.
func NestedStringSlice(obj map[string]interface{}, fields ...string) ([]string, bool) {
	val, ok := nestedFieldNoCopy(obj, fields...)
	if !ok {
		return nil, false
	}
	if m, ok := val.([]interface{}); ok {
		strSlice := make([]string, 0, len(m))
		for _, v := range m {
			if str, ok := v.(string); ok {
				strSlice = append(strSlice, str)
			} else {
				return nil, false
			}
		}
		return strSlice, true
	}
	return nil, false
}

// NestedSlice returns a deep copy of []interface{} value of a nested field.
// Returns false if value is not found or is not a []interface{}.
func NestedSlice(obj map[string]interface{}, fields ...string) ([]interface{}, bool) {
	val, ok := nestedFieldNoCopy(obj, fields...)
	if !ok {
		return nil, false
	}
	if _, ok := val.([]interface{}); ok {
		return runtime.DeepCopyJSONValue(val).([]interface{}), true
	}
	return nil, false
}

// NestedStringMap returns a copy of map[string]string value of a nested field.
// Returns false if value is not found, is not a map[string]interface{} or contains non-string values in the map.
func NestedStringMap(obj map[string]interface{}, fields ...string) (map[string]string, bool) {
	m, ok := nestedMapNoCopy(obj, fields...)
	if !ok {
		return nil, false
	}
	strMap := make(map[string]string, len(m))
	for k, v := range m {
		if str, ok := v.(string); ok {
			strMap[k] = str
		} else {
			return nil, false
		}
	}
	return strMap, true
}

// NestedMap returns a deep copy of map[string]interface{} value of a nested field.
// Returns false if value is not found or is not a map[string]interface{}.
func NestedMap(obj map[string]interface{}, fields ...string) (map[string]interface{}, bool) {
	m, ok := nestedMapNoCopy(obj, fields...)
	if !ok {
		return nil, false
	}
	return runtime.DeepCopyJSON(m), true
}

// nestedMapNoCopy returns a map[string]interface{} value of a nested field.
// Returns false if value is not found or is not a map[string]interface{}.
func nestedMapNoCopy(obj map[string]interface{}, fields ...string) (map[string]interface{}, bool) {
	val, ok := nestedFieldNoCopy(obj, fields...)
	if !ok {
		return nil, false
	}
	m, ok := val.(map[string]interface{})
	return m, ok
}

// SetNestedField sets the value of a nested field to a deep copy of the value provided.
// Returns false if value cannot be set because one of the nesting levels is not a map[string]interface{}.
func SetNestedField(obj map[string]interface{}, value interface{}, fields ...string) bool {
	return setNestedFieldNoCopy(obj, runtime.DeepCopyJSONValue(value), fields...)
}

func setNestedFieldNoCopy(obj map[string]interface{}, value interface{}, fields ...string) bool {
	m := obj
	for _, field := range fields[:len(fields)-1] {
		if val, ok := m[field]; ok {
			if valMap, ok := val.(map[string]interface{}); ok {
				m = valMap
			} else {
				return false
			}
		} else {
			newVal := make(map[string]interface{})
			m[field] = newVal
			m = newVal
		}
	}
	m[fields[len(fields)-1]] = value
	return true
}

// SetNestedStringSlice sets the string slice value of a nested field.
// Returns false if value cannot be set because one of the nesting levels is not a map[string]interface{}.
func SetNestedStringSlice(obj map[string]interface{}, value []string, fields ...string) bool {
	m := make([]interface{}, 0, len(value)) // convert []string into []interface{}
	for _, v := range value {
		m = append(m, v)
	}
	return setNestedFieldNoCopy(obj, m, fields...)
}

// SetNestedSlice sets the slice value of a nested field.
// Returns false if value cannot be set because one of the nesting levels is not a map[string]interface{}.
func SetNestedSlice(obj map[string]interface{}, value []interface{}, fields ...string) bool {
	return SetNestedField(obj, value, fields...)
}

// SetNestedStringMap sets the map[string]string value of a nested field.
// Returns false if value cannot be set because one of the nesting levels is not a map[string]interface{}.
func SetNestedStringMap(obj map[string]interface{}, value map[string]string, fields ...string) bool {
	m := make(map[string]interface{}, len(value)) // convert map[string]string into map[string]interface{}
	for k, v := range value {
		m[k] = v
	}
	return setNestedFieldNoCopy(obj, m, fields...)
}

// SetNestedMap sets the map[string]interface{} value of a nested field.
// Returns false if value cannot be set because one of the nesting levels is not a map[string]interface{}.
func SetNestedMap(obj map[string]interface{}, value map[string]interface{}, fields ...string) bool {
	return SetNestedField(obj, value, fields...)
}

// RemoveNestedField removes the nested field from the obj.
func RemoveNestedField(obj map[string]interface{}, fields ...string) {
	m := obj
	for _, field := range fields[:len(fields)-1] {
		if x, ok := m[field].(map[string]interface{}); ok {
			m = x
		} else {
			return
		}
	}
	delete(m, fields[len(fields)-1])
}

func getNestedString(obj map[string]interface{}, fields ...string) string {
	val, ok := NestedString(obj, fields...)
	if !ok {
		return ""
	}
	return val
}

func extractOwnerReference(v map[string]interface{}) metav1.OwnerReference {
	// though this field is a *bool, but when decoded from JSON, it's
	// unmarshalled as bool.
	var controllerPtr *bool
	if controller, ok := NestedBool(v, "controller"); ok {
		controllerPtr = &controller
	}
	var blockOwnerDeletionPtr *bool
	if blockOwnerDeletion, ok := NestedBool(v, "blockOwnerDeletion"); ok {
		blockOwnerDeletionPtr = &blockOwnerDeletion
	}
	return metav1.OwnerReference{
		Kind:               getNestedString(v, "kind"),
		Name:               getNestedString(v, "name"),
		APIVersion:         getNestedString(v, "apiVersion"),
		UID:                types.UID(getNestedString(v, "uid")),
		Controller:         controllerPtr,
		BlockOwnerDeletion: blockOwnerDeletionPtr,
	}
}

// UnstructuredJSONScheme is capable of converting JSON data into the Unstructured
// type, which can be used for generic access to objects without a predefined scheme.
// TODO: move into serializer/json.
var UnstructuredJSONScheme runtime.Codec = unstructuredJSONScheme{}

type unstructuredJSONScheme struct{}

func (s unstructuredJSONScheme) Decode(data []byte, _ *schema.GroupVersionKind, obj runtime.Object) (runtime.Object, *schema.GroupVersionKind, error) {
	var err error
	if obj != nil {
		err = s.decodeInto(data, obj)
	} else {
		obj, err = s.decode(data)
	}

	if err != nil {
		return nil, nil, err
	}

	gvk := obj.GetObjectKind().GroupVersionKind()
	if len(gvk.Kind) == 0 {
		return nil, &gvk, runtime.NewMissingKindErr(string(data))
	}

	return obj, &gvk, nil
}

func (unstructuredJSONScheme) Encode(obj runtime.Object, w io.Writer) error {
	switch t := obj.(type) {
	case *Unstructured:
		return json.NewEncoder(w).Encode(t.Object)
	case *UnstructuredList:
		items := make([]interface{}, 0, len(t.Items))
		for _, i := range t.Items {
			items = append(items, i.Object)
		}
		listObj := make(map[string]interface{}, len(t.Object)+1)
		for k, v := range t.Object { // Make a shallow copy
			listObj[k] = v
		}
		listObj["items"] = items
		return json.NewEncoder(w).Encode(listObj)
	case *runtime.Unknown:
		// TODO: Unstructured needs to deal with ContentType.
		_, err := w.Write(t.Raw)
		return err
	default:
		return json.NewEncoder(w).Encode(t)
	}
}

func (s unstructuredJSONScheme) decode(data []byte) (runtime.Object, error) {
	type detector struct {
		Items gojson.RawMessage
	}
	var det detector
	if err := json.Unmarshal(data, &det); err != nil {
		return nil, err
	}

	if det.Items != nil {
		list := &UnstructuredList{}
		err := s.decodeToList(data, list)
		return list, err
	}

	// No Items field, so it wasn't a list.
	unstruct := &Unstructured{}
	err := s.decodeToUnstructured(data, unstruct)
	return unstruct, err
}

func (s unstructuredJSONScheme) decodeInto(data []byte, obj runtime.Object) error {
	switch x := obj.(type) {
	case *Unstructured:
		return s.decodeToUnstructured(data, x)
	case *UnstructuredList:
		return s.decodeToList(data, x)
	case *runtime.VersionedObjects:
		o, err := s.decode(data)
		if err == nil {
			x.Objects = []runtime.Object{o}
		}
		return err
	default:
		return json.Unmarshal(data, x)
	}
}

func (unstructuredJSONScheme) decodeToUnstructured(data []byte, unstruct *Unstructured) error {
	m := make(map[string]interface{})
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	unstruct.Object = m

	return nil
}

func (s unstructuredJSONScheme) decodeToList(data []byte, list *UnstructuredList) error {
	type decodeList struct {
		Items []gojson.RawMessage
	}

	var dList decodeList
	if err := json.Unmarshal(data, &dList); err != nil {
		return err
	}

	if err := json.Unmarshal(data, &list.Object); err != nil {
		return err
	}

	// For typed lists, e.g., a PodList, API server doesn't set each item's
	// APIVersion and Kind. We need to set it.
	listAPIVersion := list.GetAPIVersion()
	listKind := list.GetKind()
	itemKind := strings.TrimSuffix(listKind, "List")

	delete(list.Object, "items")
	list.Items = make([]Unstructured, 0, len(dList.Items))
	for _, i := range dList.Items {
		unstruct := &Unstructured{}
		if err := s.decodeToUnstructured([]byte(i), unstruct); err != nil {
			return err
		}
		// This is hacky. Set the item's Kind and APIVersion to those inferred
		// from the List.
		if len(unstruct.GetKind()) == 0 && len(unstruct.GetAPIVersion()) == 0 {
			unstruct.SetKind(itemKind)
			unstruct.SetAPIVersion(listAPIVersion)
		}
		list.Items = append(list.Items, *unstruct)
	}
	return nil
}

// UnstructuredObjectConverter is an ObjectConverter for use with
// Unstructured objects. Since it has no schema or type information,
// it will only succeed for no-op conversions. This is provided as a
// sane implementation for APIs that require an object converter.
type UnstructuredObjectConverter struct{}

func (UnstructuredObjectConverter) Convert(in, out, context interface{}) error {
	unstructIn, ok := in.(*Unstructured)
	if !ok {
		return fmt.Errorf("input type %T in not valid for unstructured conversion", in)
	}

	unstructOut, ok := out.(*Unstructured)
	if !ok {
		return fmt.Errorf("output type %T in not valid for unstructured conversion", out)
	}

	// maybe deep copy the map? It is documented in the
	// ObjectConverter interface that this function is not
	// guaranteed to not mutate the input. Or maybe set the input
	// object to nil.
	unstructOut.Object = unstructIn.Object
	return nil
}

func (UnstructuredObjectConverter) ConvertToVersion(in runtime.Object, target runtime.GroupVersioner) (runtime.Object, error) {
	if kind := in.GetObjectKind().GroupVersionKind(); !kind.Empty() {
		gvk, ok := target.KindForGroupVersionKinds([]schema.GroupVersionKind{kind})
		if !ok {
			// TODO: should this be a typed error?
			return nil, fmt.Errorf("%v is unstructured and is not suitable for converting to %q", kind, target)
		}
		in.GetObjectKind().SetGroupVersionKind(gvk)
	}
	return in, nil
}

func (UnstructuredObjectConverter) ConvertFieldLabel(version, kind, label, value string) (string, string, error) {
	return "", "", errors.New("unstructured cannot convert field labels")
}
