/*
 * Copyright (C) 2020 ObjectBox Ltd. All rights reserved.
 * https://objectbox.io
 *
 * This file is part of ObjectBox Generator.
 *
 * ObjectBox Generator is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 * ObjectBox Generator is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with ObjectBox Generator.  If not, see <http://www.gnu.org/licenses/>.
 */

package templates

import (
	"text/template"
)

// BindingTemplate is used to generated the binding code
var BindingTemplate = template.Must(template.New("binding").Funcs(funcMap).Parse(
	`// Code generated by ObjectBox; DO NOT EDIT. 
// Learn more about defining entities and generating this file - visit https://golang.objectbox.io/entity-annotations

{{define "property-getter-with-converter-val"}}{{/* used in Load*/}}
	{{- if .Converter}} prop{{.Name}}
	{{- else}} {{template "property-getter" .}}
	{{- end}}
{{- end -}}
{{define "property-getter"}}{{/* used in Load*/}}
	{{- if .CastOnWrite}}{{.CastOnWrite}}({{end}}
		{{- if eq .FbType "UOffsetT"}} fbutils.Get{{.ObTypeString}}{{if .GoField.IsPointer}}Ptr{{end}}Slot(table, {{.ModelProperty.FbvTableOffset}})
    	{{- else}} fbutils.Get{{.GoType | StringTitle}}{{if .GoField.IsPointer}}Ptr{{end}}Slot(table, {{.ModelProperty.FbvTableOffset}})
    	{{- end}}
	{{- if .CastOnWrite}}){{end}}
{{- end -}}

{{define "property-access"}}{{/* used in Flatten*/ -}}
	{{- if .Converter}} {{if .GoField.IsPointer}}*{{end}}prop{{.Name}}
	{{- else if .CastOnRead}}{{.CastOnRead}}({{if .GoField.IsPointer}}*{{end}}obj.{{.Path}})
	{{- else}}{{if .GoField.IsPointer}}*{{end}}obj.{{.Path}}{{end}}
{{- end -}}


package {{.Binding.Package.Name}}

import (
	"errors"
	"github.com/google/flatbuffers/go"
	"github.com/objectbox/objectbox-go/objectbox"
	"github.com/objectbox/objectbox-go/objectbox/fbutils"
	{{range $alias, $path := .Binding.Imports -}}
		{{if not (eq $alias $path)}}{{$alias}}{{end}} "{{$path}}"
	{{end}}
)

{{range $entity := .Model.EntitiesWithMeta -}}
{{$entityNameCamel := $entity.Name | StringCamel -}}
type {{$entityNameCamel}}_EntityInfo struct {
	objectbox.Entity
	Uid uint64
}

var {{$entity.Name}}Binding = {{$entityNameCamel}}_EntityInfo {
	Entity: objectbox.Entity{
		Id: {{$entity.Id.GetId}},
	}, 
	Uid: {{$entity.Id.GetUid}},
}

// {{$entity.Name}}_ contains type-based Property helpers to facilitate some common operations such as Queries. 
var {{$entity.Name}}_ = struct {
	{{range $property := $entity.Properties -}}
    	{{$property.Meta.Name}} *objectbox.{{with $property.RelationTarget}}RelationToOne{{else}}Property{{$property.Meta.GoType | TypeIdentifier}}{{end}}
    {{end -}}
	{{range $relation := $entity.Relations -}}
    	{{$relation.Name}} *objectbox.RelationToMany
	{{end -}}
}{
	{{range $property := $entity.Properties -}}
    {{$property.Meta.Name}}: &objectbox.
		{{- with $property.RelationTarget}}RelationToOne{
			Property:
		{{- else}}Property{{$property.Meta.GoType | TypeIdentifier}}{
			BaseProperty:
		{{- end -}} 
		&objectbox.BaseProperty{
			Id: {{$property.Id.GetId}},
			Entity: &{{$entity.Name}}Binding.Entity,
		},{{with $property.RelationTarget}}
		Target: &{{.}}Binding.Entity,{{end}}
	},
    {{end -}}
	{{range $relation := $entity.Relations -}}
    	{{$relation.Name}}: &objectbox.RelationToMany{
			Id: {{$relation.Id.GetId}},
			Source: &{{$entity.Name}}Binding.Entity,
			Target: &{{$relation.Target.Name}}Binding.Entity,
		},
    {{end -}}
}

// GeneratorVersion is called by ObjectBox to verify the compatibility of the generator used to generate this code	
func ({{$entityNameCamel}}_EntityInfo) GeneratorVersion() int {
	return {{$.GeneratorVersion}}
}

// AddToModel is called by ObjectBox during model build
func ({{$entityNameCamel}}_EntityInfo) AddToModel(model *objectbox.Model) {
    model.Entity("{{$entity.Name}}", {{$entity.Id.GetId}}, {{$entity.Id.GetUid}})
    {{with $entity.Flags -}}
		model.EntityFlags({{.}})
	{{end -}}
    {{range $property := $entity.Properties -}}
    model.Property("{{$property.Name}}", {{$property.Type}}, {{$property.Id.GetId}}, {{$property.Id.GetUid}})
    {{with $property.Flags -}}
		model.PropertyFlags({{.}})
	{{end -}}
	{{if $property.RelationTarget}}model.PropertyRelation("{{$property.RelationTarget}}", {{$property.IndexId.GetId}}, {{$property.IndexId.GetUid}})
	{{else if $property.IndexId}}model.PropertyIndex({{$property.IndexId.GetId}}, {{$property.IndexId.GetUid}})
    {{end -}}
    {{end -}}
    model.EntityLastPropertyId({{$entity.LastPropertyId.GetId}}, {{$entity.LastPropertyId.GetUid}})
	{{range $relation := $entity.Relations -}}
    model.Relation({{$relation.Id.GetId}}, {{$relation.Id.GetUid}}, {{$relation.Target.Name}}Binding.Id, {{$relation.Target.Name}}Binding.Uid)
    {{end -}}
}

// GetId is called by ObjectBox during Put operations to check for existing ID on an object
func ({{$entityNameCamel}}_EntityInfo) GetId(object interface{}) (uint64, error) {
	{{- if $.ByValue}}
		if obj, ok := object.(*{{$entity.Name}}); ok {
			return {{$entity.IdProperty.Meta.TplReadValue "obj" ""}}
		} else {
			return {{$entity.IdProperty.Meta.TplReadValue "object" "val-cast"}}
		}
	{{- else -}}
		return {{$entity.IdProperty.Meta.TplReadValue "object" "ptr-cast"}}
	{{- end}}
}

// SetId is called by ObjectBox during Put to update an ID on an object that has just been inserted
func ({{$entityNameCamel}}_EntityInfo) SetId(object interface{}, id uint64) error {
	{{- if $.ByValue}}
		if obj, ok := object.(*{{$entity.Name}}); ok {
			{{$entity.IdProperty.Meta.TplSetAndReturn "obj" "" "id"}}
		} else {
			// NOTE while this can't update, it will at least behave consistently (panic in case of a wrong type)
			_ = object.({{$entity.Name}}).{{$entity.IdProperty.Meta.Path}}
			return nil
		}
	{{- else -}}
		{{$entity.IdProperty.Meta.TplSetAndReturn "object" "ptr-cast" "id"}}
	{{- end}}
}

// PutRelated is called by ObjectBox to put related entities before the object itself is flattened and put
func ({{$entityNameCamel}}_EntityInfo) PutRelated(ob *objectbox.ObjectBox, object interface{}, id uint64) error {
	{{- block "put-relations" $entity}}
	{{- range $field := .Meta.Fields}}
		{{- if $field.StandaloneRelation}}
			{{- if $field.IsLazyLoaded}} if object.(*{{$field.Entity.Name}}).{{$field.Path}} != nil { // lazy-loaded relations without {{$field.Entity.Name}}Box::Fetch{{$field.Name}}() called are nil {{end}}  
			if err := BoxFor{{$field.Entity.Name}}(ob).RelationReplace({{.Entity.Name}}_.{{$field.Name}}, id, object, object.(*{{$field.Entity.Name}}).{{$field.Path}}); err != nil {
				return err
			}
			{{if $field.IsLazyLoaded}} }
   			{{end}}
		{{- else if $field.Property}}
			{{- if and (not $field.Property.IsBasicType) $field.Property.ModelProperty.RelationTarget}}
			if rel := {{if not $field.IsPointer}}&{{end}}object.(*{{$field.Entity.Name}}).{{$field.Path}}; rel != nil {
				if rId, err := {{$field.Property.ModelProperty.RelationTarget}}Binding.GetId(rel); err != nil {
					return err
				} else if rId == 0 {
					// NOTE Put/PutAsync() has a side-effect of setting the rel.ID
					if _, err := BoxFor{{$field.Property.ModelProperty.RelationTarget}}(ob).Put(rel); err != nil {
						return err
					}
				}
			}
			{{- end}}
		{{- else}}{{/* recursively visit fields in embedded structs */}}{{template "put-relations" $field}}
		{{- end}}
	{{- end}}{{end}}
	return nil
}

// Flatten is called by ObjectBox to transform an object to a FlatBuffer
func ({{$entityNameCamel}}_EntityInfo) Flatten(object interface{}, fbb *flatbuffers.Builder, id uint64) error {
    {{if $entity.Meta.HasNonIdProperty -}}
		{{- if not $.ByValue}}obj := object.(*{{$entity.Name}}) 
		{{- else -}}
		var obj *{{$entity.Name}}
		if objPtr, ok := object.(*{{$entity.Name}}); ok {
			obj = objPtr 
		} else {
			objVal := object.({{$entity.Name}})
			obj = &objVal
		}
		{{end}}
	{{- end -}}
	
	{{- range $property := $entity.Properties}}{{if and $property.Meta.Converter (not (eq $property.Name $entity.IdProperty.Name))}}
	var prop{{$property.Name}} {{$property.Meta.AnnotatedType}}
	{{if $property.Meta.GoField.IsPointer}}if obj.{{$property.Meta.Path}} != nil {{end}} { 
		var err error
		prop{{$property.Name}}, err = {{$property.Meta.Converter}}ToDatabaseValue(obj.{{$property.Meta.Path}})
		if err != nil {
			return errors.New("converter {{$property.Meta.Converter}}ToDatabaseValue() failed on {{$entity.Name}}.{{$property.Meta.Path}}: " + err.Error())
		}
	}
	{{end}}{{end}}

    {{- range $property := $entity.Properties}}{{if eq $property.Meta.FbType "UOffsetT"}}
	{{if $property.Meta.GoField.IsPointer}}
	var offset{{$property.Meta.Name}} flatbuffers.UOffsetT
	if obj.{{$property.Meta.Path}} != nil {
	{{else}}var {{end -}}
	offset{{$property.Meta.Name}} = fbutils.Create{{$property.Meta.ObTypeString}}Offset(fbb, {{template "property-access" $property.Meta}})
	{{- if $property.Meta.GoField.IsPointer -}} } {{- end}}
	{{- end}}{{end}}

	{{- block "store-relations" $entity}}
	{{- range $field := .Meta.Fields}}
		{{if $field.Property}}
			{{- if $field.Property.ModelProperty.RelationTarget}}
				{{- if $field.Property.IsBasicType}}{{/* manual relation links (just ID) */}}
					var rId{{$field.Name}} = {{template "property-access" $field.Property}}
				{{- else}}
					var rId{{$field.Name}} uint64
					if rel := {{if not $field.IsPointer}}&{{end}}obj.{{$field.Path}}; rel != nil {
						if rId, err := {{$field.Property.ModelProperty.RelationTarget}}Binding.GetId(rel); err != nil {
							return err
						} else {
							rId{{$field.Name}} = rId
						}
					}
				{{- end}}
			{{- end}}
		{{- else}}{{/* recursively visit fields in embedded structs */}}{{template "store-relations" $field}}
		{{end}}
	{{end}}{{end}}

    // build the FlatBuffers object
    fbb.StartObject({{$entity.LastPropertyId.GetId}})
	{{- if $entity.IdProperty.Meta.GoField.HasPointersInPath }}{{/* when Id property's path (embedded) contains pointers, make sure it's always set */}} 
		fbutils.Set{{$entity.IdProperty.Meta.FbType}}Slot(fbb, {{$entity.IdProperty.FbSlot}}, id) 
	{{- end}}
	{{- block "fields-setter" $entity -}}
		{{- range $field := .Meta.Fields}}
			{{- if $field.IsPointer}}
			if obj.{{$field.Path}} != nil { {{- end -}}
			{{with $field.Property}}{{if or (not .ModelProperty.IsIdProperty) (not .GoField.HasPointersInPath) }} 
				fbutils.Set{{.FbType}}Slot(fbb, {{.ModelProperty.FbSlot}},
				{{- if .ModelProperty.RelationTarget}}rId{{.Name}})
				{{- else if eq .FbType "UOffsetT"}} offset{{.Name}})
				{{- else if .ModelProperty.IsIdProperty}} id)
				{{- else -}}
					{{- if or (eq .GoType "int") (eq .GoType "uint") }} {{.GoType}}64( {{end}} 
					{{- template "property-access" . -}})
					{{- if or (eq .GoType "int") (eq .GoType "uint")}} ) {{end}}
				{{- end}}{{end}}
			{{- else}}{{template "fields-setter" $field}}{{end -}}
			{{- if $field.IsPointer -}} } {{- end -}}
		{{- end -}}
	{{end}}
	return nil
}

// Load is called by ObjectBox to load an object from a FlatBuffer 
func ({{$entityNameCamel}}_EntityInfo) Load(ob *objectbox.ObjectBox, bytes []byte) (interface{}, error) {
	if len(bytes) == 0 { // sanity check, should "never" happen
		return nil, errors.New("can't deserialize an object of type '{{$entity.Name}}' - no data received")
	}

	var table = &flatbuffers.Table{
		Bytes: bytes,
		Pos:   flatbuffers.GetUOffsetT(bytes),
	}

	{{if not $entity.IdProperty.Meta.Converter}}
	var prop{{$entity.IdProperty.Name}} = table.Get{{$entity.IdProperty.Meta.GoType | StringTitle}}Slot({{$entity.IdProperty.FbvTableOffset}}, 0)
	{{end -}}

	{{range $property := $entity.Properties}}{{if $property.Meta.Converter}}
	prop{{$property.Name}}, err := {{$property.Meta.Converter}}ToEntityProperty({{template "property-getter" $property.Meta}})
	if err != nil {
		return nil, errors.New("converter {{$property.Meta.Converter}}ToEntityProperty() failed on {{$entity.Name}}.{{$property.Meta.Path}}: " + err.Error())
	}
	{{end}}{{end}}
	
	{{- block "load-relations" $entity}}
	{{- range $field := .Meta.Fields}}
		{{if $field.StandaloneRelation -}}
			{{if not $field.IsLazyLoaded -}}
			var rel{{$field.Name}} {{$field.Type}} 
			if rIds, err := BoxFor{{$field.Entity.Name}}(ob).RelationIds({{.Entity.Name}}_.{{$field.Name}}, prop{{.Entity.ModelEntity.IdProperty.Name}}); err != nil {
				return nil, err
			} else if rSlice, err := BoxFor{{$field.StandaloneRelation.Target.Name}}(ob).GetManyExisting(rIds...); err != nil {
				return nil, err
			} else {
				rel{{$field.Name}} = rSlice
			}
			{{- end -}} {{/* see Fetch* for lazy loaded relations */}}
		{{else if $field.Property -}}
			{{- if and (not $field.Property.IsBasicType) $field.Property.ModelProperty.RelationTarget }}
			var rel{{$field.Name}} *{{$field.Type}}
			if rId := {{template "property-getter-with-converter-val" $field.Property}}; {{if $field.Property.GoField.IsPointer}}rId != nil && *{{end}}rId > 0 {
				if rObject, err := BoxFor{{$field.Property.ModelProperty.RelationTarget}}(ob).Get({{if $field.Property.GoField.IsPointer}}*{{end}}rId); err != nil {
					return nil, err 
				{{if not $field.IsPointer -}}
				} else if rObject == nil {
					rel{{$field.Name}} = &{{$field.Type}}{}
				{{end -}}
				} else {
					rel{{$field.Name}} = rObject
				}
			{{if not $field.IsPointer -}} 
			} else {
				rel{{$field.Name}} = &{{$field.Type}}{}
			{{end -}}
			}
			{{- end}}
		{{else}}{{/* recursively visit fields in embedded structs */}}{{template "load-relations" $field}}
		{{- end}}
	{{end}}{{end}}

	return &{{$entity.Name}}{
	{{- block "fields-initializer" $entity}}
		{{- range $field := .Meta.Fields}}
			{{$field.Name}}: 
				{{- if $field.StandaloneRelation}}
					{{- if $field.IsLazyLoaded}}nil, // use {{$field.Entity.Name}}Box::Fetch{{$field.Name}}() to fetch this lazy-loaded relation
					{{- else}}rel{{$field.Name}}
					{{- end}}
				{{- else if $field.Property}}
					{{- if and (not $field.Property.IsBasicType) $field.Property.ModelProperty.RelationTarget}}{{if not $field.IsPointer}}*{{end}}rel{{$field.Name}}
					{{- else if $field.Property.ModelProperty.IsIdProperty}} prop{{$field.Property.Name}}
					{{- else}}{{template "property-getter-with-converter-val" $field.Property}}
					{{- end}}
				{{- else}}{{if $field.IsPointer}}&{{end}}{{$field.Type}}{ {{template "fields-initializer" $field}} }
				{{- end}},
		{{- end}}
	{{end}}
	}, nil
}

// MakeSlice is called by ObjectBox to construct a new slice to hold the read objects  
func ({{$entityNameCamel}}_EntityInfo) MakeSlice(capacity int) interface{} {
	return make([]{{if not $.ByValue}}*{{end}}{{$entity.Name}}, 0, capacity)
}

// AppendToSlice is called by ObjectBox to fill the slice of the read objects
func ({{$entityNameCamel}}_EntityInfo) AppendToSlice(slice interface{}, object interface{}) interface{} {
	if object == nil {
		return append(slice.([]{{if not $.ByValue}}*{{end}}{{$entity.Name}}), {{if $.ByValue}}{{$entity.Name}}{}{{else}}nil{{end}})
	}
	return append(slice.([]{{if not $.ByValue}}*{{end}}{{$entity.Name}}), {{if $.ByValue}}*{{end}}object.(*{{$entity.Name}}))
}

// Box provides CRUD access to {{$entity.Name}} objects
type {{$entity.Name}}Box struct {
	*objectbox.Box
}

// BoxFor{{$entity.Name}} opens a box of {{$entity.Name}} objects 
func BoxFor{{$entity.Name}}(ob *objectbox.ObjectBox) *{{$entity.Name}}Box {
	return &{{$entity.Name}}Box{
		Box: ob.InternalBox({{$entity.Id.GetId}}),
	}
}

// Put synchronously inserts/updates a single object.
// In case the {{$entity.IdProperty.Meta.Path}} is not specified, it would be assigned automatically (auto-increment).
// When inserting, the {{$entity.Name}}.{{$entity.IdProperty.Meta.Path}} property on the passed object will be assigned the new ID as well.
func (box *{{$entity.Name}}Box) Put(object *{{$entity.Name}}) (uint64, error) {
	return box.Box.Put(object)
}

// Insert synchronously inserts a single object. As opposed to Put, Insert will fail if given an ID that already exists.
// In case the {{$entity.IdProperty.Meta.Path}} is not specified, it would be assigned automatically (auto-increment).
// When inserting, the {{$entity.Name}}.{{$entity.IdProperty.Meta.Path}} property on the passed object will be assigned the new ID as well.
func (box *{{$entity.Name}}Box) Insert(object *{{$entity.Name}}) (uint64, error) {
	return box.Box.Insert(object)
}

// Update synchronously updates a single object.
// As opposed to Put, Update will fail if an object with the same ID is not found in the database.
func (box *{{$entity.Name}}Box) Update(object *{{$entity.Name}}) error {
	return box.Box.Update(object)
}

// PutAsync asynchronously inserts/updates a single object.
// Deprecated: use box.Async().Put() instead
func (box *{{$entity.Name}}Box) PutAsync(object *{{$entity.Name}}) (uint64, error) {
	return box.Box.PutAsync(object)
}

// PutMany inserts multiple objects in single transaction.
// In case {{$entity.IdProperty.Meta.Path}}s are not set on the objects, they would be assigned automatically (auto-increment).
// 
// Returns: IDs of the put objects (in the same order).
// When inserting, the {{$entity.Name}}.{{$entity.IdProperty.Meta.Path}} property on the objects in the slice will be assigned the new IDs as well.
//
// Note: In case an error occurs during the transaction, some of the objects may already have the {{$entity.Name}}.{{$entity.IdProperty.Meta.Path}} assigned    
// even though the transaction has been rolled back and the objects are not stored under those IDs.
//
// Note: The slice may be empty or even nil; in both cases, an empty IDs slice and no error is returned.
func (box *{{$entity.Name}}Box) PutMany(objects []{{if not $.ByValue}}*{{end}}{{$entity.Name}}) ([]uint64, error) {
	return box.Box.PutMany(objects)
}

// Get reads a single object.
//
// Returns nil (and no error) in case the object with the given ID doesn't exist.
func (box *{{$entity.Name}}Box) Get(id uint64) (*{{$entity.Name}}, error) {
	object, err := box.Box.Get(id)
	if err != nil {
		return nil, err
	} else if object == nil {
		return nil, nil
	}
	return object.(*{{$entity.Name}}), nil
}

// GetMany reads multiple objects at once.
// If any of the objects doesn't exist, its position in the return slice is {{if $.ByValue}}an empty object{{else}}nil{{end}}
func (box *{{$entity.Name}}Box) GetMany(ids ...uint64) ([]{{if not $.ByValue}}*{{end}}{{$entity.Name}}, error) {
	objects, err := box.Box.GetMany(ids...)
	if err != nil {
		return nil, err
	}
	return objects.([]{{if not $.ByValue}}*{{end}}{{$entity.Name}}), nil
}

// GetManyExisting reads multiple objects at once, skipping those that do not exist.
func (box *{{$entity.Name}}Box) GetManyExisting(ids ...uint64) ([]{{if not $.ByValue}}*{{end}}{{$entity.Name}}, error) {
	objects, err := box.Box.GetManyExisting(ids...)
	if err != nil {
		return nil, err
	}
	return objects.([]{{if not $.ByValue}}*{{end}}{{$entity.Name}}), nil
}

// GetAll reads all stored objects
func (box *{{$entity.Name}}Box) GetAll() ([]{{if not $.ByValue}}*{{end}}{{$entity.Name}}, error) {
	objects, err := box.Box.GetAll()
	if err != nil {
		return nil, err
	}
	return objects.([]{{if not $.ByValue}}*{{end}}{{$entity.Name}}), nil
}

{{- block "fetch-related" $entity}}
{{- range $field := .Meta.Fields}}
	{{if .StandaloneRelation}}
		{{- if .IsLazyLoaded -}}
			// Fetch{{.Name}} reads target objects for relation {{.Entity.Name}}::{{.Name}}.
			// It will "GetManyExisting()" all related {{.StandaloneRelation.Target.Name}} objects for each source object
			// and set sourceObject.{{.Name}} to the slice of related objects, as currently stored in DB.
			func (box *{{.Entity.Name}}Box) Fetch{{.Name}}(sourceObjects ...*{{.Entity.Name}}) error {
				var slices = make([]{{.Type}}, len(sourceObjects))
				err := box.ObjectBox.RunInReadTx(func() error {
					// collect slices before setting the source objects' fields
					// this keeps all the sourceObjects untouched in case there's an error during any of the requests
					for k, object := range sourceObjects {
						{{if .Entity.ModelEntity.IdProperty.Meta.Converter -}}
						sourceId, err := {{.Entity.ModelEntity.IdProperty.Meta.Converter}}ToDatabaseValue(object.{{.Entity.ModelEntity.IdProperty.Meta.Path}})
						if err != nil {
							return err
						}
						{{end -}}
						rIds, err := box.RelationIds({{.Entity.Name}}_.{{.Name}}, {{with .Entity.ModelEntity.IdProperty}} {{if .Meta.Converter}}sourceId{{else}}object.{{.Meta.Path}}{{end}}{{end}})
						if err == nil {
						    slices[k], err = BoxFor{{.StandaloneRelation.Target.Name}}(box.ObjectBox).GetManyExisting(rIds...)
						}
						if err != nil {
							return err
						}
					}
					return nil
                })

				if err == nil {  // update the field on all objects if we got all slices 
					for k := range sourceObjects {
						sourceObjects[k].{{.Name}} = slices[k]
					}
				}
				return err
			}
		{{end}}
	{{- else if not .Property}}{{/* recursively visit fields in embedded structs */}}{{template "fetch-related" $field}}
	{{- end}}
{{- end}}{{end}}

// Remove deletes a single object
func (box *{{$entity.Name}}Box) Remove(object *{{$entity.Name}}) error {
	return box.Box.Remove(object)
}

// RemoveMany deletes multiple objects at once.
// Returns the number of deleted object or error on failure.
// Note that this method will not fail if an object is not found (e.g. already removed).
// In case you need to strictly check whether all of the objects exist before removing them,
// you can execute multiple box.Contains() and box.Remove() inside a single write transaction.
func (box *{{$entity.Name}}Box) RemoveMany(objects ...*{{$entity.Name}}) (uint64, error) {
	var ids = make([]uint64, len(objects))
	{{- if $entity.IdProperty.Meta.Converter}}
	var err error{{end}}
	for k, object := range objects {
		{{if $entity.IdProperty.Meta.Converter -}}
			ids[k], err = {{$entity.IdProperty.Meta.TplReadValue "object" ""}}
			if err != nil {
				return 0, errors.New("converter {{$entity.IdProperty.Meta.Converter}}ToDatabaseValue() failed on {{$entity.Name}}.{{$entity.IdProperty.Meta.Path}}: " + err.Error())
			}
		{{else -}}
			ids[k] = {{with $entity.IdProperty -}}
				{{- if not (eq .Meta.GoType "uint64")}} uint64( {{end -}}
				object.{{.Meta.Path}}
				{{- if not (eq .Meta.GoType "uint64")}} ) {{end -}}
			{{- end}}
		{{end -}}
	}
	return box.Box.RemoveIds(ids...)
}

// Creates a query with the given conditions. Use the fields of the {{$entity.Name}}_ struct to create conditions.
// Keep the *{{$entity.Name}}Query if you intend to execute the query multiple times.
// Note: this function panics if you try to create illegal queries; e.g. use properties of an alien type.
// This is typically a programming error. Use QueryOrError instead if you want the explicit error check.
func (box *{{$entity.Name}}Box) Query(conditions ...objectbox.Condition) *{{$entity.Name}}Query {
	return &{{$entity.Name}}Query{
		box.Box.Query(conditions...),
	}
}

// Creates a query with the given conditions. Use the fields of the {{$entity.Name}}_ struct to create conditions.
// Keep the *{{$entity.Name}}Query if you intend to execute the query multiple times.
func (box *{{$entity.Name}}Box) QueryOrError(conditions ...objectbox.Condition) (*{{$entity.Name}}Query, error) {
	if query, err := box.Box.QueryOrError(conditions...); err != nil {
		return nil, err
	} else {
		return &{{$entity.Name}}Query{query}, nil
	}
}

// Async provides access to the default Async Box for asynchronous operations. See {{$entity.Name}}AsyncBox for more information.
func (box *{{$entity.Name}}Box) Async() *{{$entity.Name}}AsyncBox {
	return &{{$entity.Name}}AsyncBox{AsyncBox: box.Box.Async()}
}

// {{$entity.Name}}AsyncBox provides asynchronous operations on {{$entity.Name}} objects.
//
// Asynchronous operations are executed on a separate internal thread for better performance.
//
// There are two main use cases:
//
// 1) "execute & forget:" you gain faster put/remove operations as you don't have to wait for the transaction to finish.
//
// 2) Many small transactions: if your write load is typically a lot of individual puts that happen in parallel,
// this will merge small transactions into bigger ones. This results in a significant gain in overall throughput.
//
// In situations with (extremely) high async load, an async method may be throttled (~1ms) or delayed up to 1 second.
// In the unlikely event that the object could still not be enqueued (full queue), an error will be returned.
//
// Note that async methods do not give you hard durability guarantees like the synchronous Box provides.
// There is a small time window in which the data may not have been committed durably yet.
type {{$entity.Name}}AsyncBox struct {
	*objectbox.AsyncBox
}

// AsyncBoxFor{{$entity.Name}} creates a new async box with the given operation timeout in case an async queue is full.
// The returned struct must be freed explicitly using the Close() method.
// It's usually preferable to use {{$entity.Name}}Box::Async() which takes care of resource management and doesn't require closing.
func AsyncBoxFor{{$entity.Name}}(ob *objectbox.ObjectBox, timeoutMs uint64) *{{$entity.Name}}AsyncBox {
	var async, err = objectbox.NewAsyncBox(ob, {{$entity.Id.GetId}}, timeoutMs)
	if err != nil {
		panic("Could not create async box for entity ID {{$entity.Id.GetId}}: %s" + err.Error())
	}
	return &{{$entity.Name}}AsyncBox{AsyncBox: async}
}

// Put inserts/updates a single object asynchronously.
// When inserting a new object, the {{$entity.IdProperty.Meta.Path}} property on the passed object will be assigned the new ID the entity would hold
// if the insert is ultimately successful. The newly assigned ID may not become valid if the insert fails.
func (asyncBox *{{$entity.Name}}AsyncBox) Put(object *{{$entity.Name}}) (uint64, error) {
	return asyncBox.AsyncBox.Put(object)
}

// Insert a single object asynchronously.
// The {{$entity.IdProperty.Meta.Path}} property on the passed object will be assigned the new ID the entity would hold if the insert is ultimately
// successful. The newly assigned ID may not become valid if the insert fails.
// Fails silently if an object with the same ID already exists (this error is not returned).
func (asyncBox *{{$entity.Name}}AsyncBox) Insert(object *{{$entity.Name}})  (id uint64, err error) {
	return asyncBox.AsyncBox.Insert(object)
}

// Update a single object asynchronously.
// The object must already exists or the update fails silently (without an error returned).
func (asyncBox *{{$entity.Name}}AsyncBox) Update(object *{{$entity.Name}}) error {
	return asyncBox.AsyncBox.Update(object)
}

// Remove deletes a single object asynchronously.
func (asyncBox *{{$entity.Name}}AsyncBox) Remove(object *{{$entity.Name}}) error {
	return asyncBox.AsyncBox.Remove(object)
}

// Query provides a way to search stored objects
//
// For example, you can find all {{$entity.Name}} which {{$entity.IdProperty.Meta.Name}} is either 42 or 47:
// 		box.Query({{$entity.Name}}_.{{$entity.IdProperty.Meta.Name}}.In(42, 47)).Find()
type {{$entity.Name}}Query struct {
	*objectbox.Query
}

// Find returns all objects matching the query
func (query *{{$entity.Name}}Query) Find() ([]{{if not $.ByValue}}*{{end}}{{$entity.Name}}, error) {
	objects, err := query.Query.Find()
	if err != nil {
		return nil, err
	}
	return objects.([]{{if not $.ByValue}}*{{end}}{{$entity.Name}}), nil
}

// Offset defines the index of the first object to process (how many objects to skip)
func (query *{{$entity.Name}}Query) Offset(offset uint64) *{{$entity.Name}}Query {
	query.Query.Offset(offset)
	return query
}

// Limit sets the number of elements to process by the query
func (query *{{$entity.Name}}Query) Limit(limit uint64) *{{$entity.Name}}Query {
	query.Query.Limit(limit)
	return query
}
{{end -}}`))
