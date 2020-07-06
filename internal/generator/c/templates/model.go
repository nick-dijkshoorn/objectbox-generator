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

// ModelTemplate is used to generate the model initialization code
var ModelTemplate = template.Must(template.New("model").Funcs(funcMap).Parse(
	`// Code generated by ObjectBox; DO NOT EDIT.

#pragma once

#ifdef __cplusplus
#include <cstdbool>
#include <cstdint>
extern "C" {
#else
#include <stdbool.h>
#include <stdint.h>
#endif
#include "objectbox.h"

/// Initializes an ObjectBox model for all entities. 
/// The returned pointer may be NULL if the allocation failed. If the returned model is not NULL, you should check if   
/// any error occurred by calling obx_model_error_code() and/or obx_model_error_message(). If an error occurred, you're
/// responsible for freeing the resources by calling obx_model_free().
/// In case there was no error when setting the model up (i.e. obx_model_error_code() returned 0), you may configure 
/// OBX_store_options with the model by calling obx_opt_model() and subsequently opening a store with obx_store_open().
/// As soon as you call obx_store_open(), the model pointer is consumed and MUST NOT be freed manually.
static inline OBX_model* create_obx_model() {
    OBX_model* model = obx_model();
    if (!model) return NULL;
	{{range $entity := .Model.Entities}}
	obx_model_entity(model, "{{$entity.Name}}", {{$entity.Id.GetId}}, {{$entity.Id.GetUid}});
	{{range $property := $entity.Properties -}}
	obx_model_property(model, "{{$property.Name}}", OBXPropertyType_{{PropTypeName $property.Type}}, {{$property.Id.GetId}}, {{$property.Id.GetUid}});
	{{with $property.Flags}}obx_model_property_flags(model, {{CorePropFlags .}});
	{{end -}}
	{{if $property.RelationTarget}}obx_model_property_relation(model, "{{$property.RelationTarget}}", {{$property.IndexId.GetId}}, {{$property.IndexId.GetUid}});
	{{else if $property.IndexId}}obx_model_property_index_id(model, {{$property.IndexId.GetId}}, {{$property.IndexId.GetUid}});
	{{end -}}
	{{range $relation := $entity.Relations -}}
    obx_model_relation(model, {{$relation.Id.GetId}}, {{$relation.Id.GetUid}}, {{$relation.Target.Id.GetId}}, {{$relation.Target.Id.GetUid}});
	{{end -}}
	{{end -}}
	obx_model_entity_last_property_id(model, {{$entity.LastPropertyId.GetId}}, {{$entity.LastPropertyId.GetUid}});
	{{end}}
	obx_model_last_entity_id(model, {{.Model.LastEntityId.GetId}}, {{.Model.LastEntityId.GetUid}});
	{{- if .Model.LastIndexId}}
	obx_model_last_index_id(model, {{.Model.LastIndexId.GetId}}, {{.Model.LastIndexId.GetUid}});
	{{- end}}
	{{- if .Model.LastRelationId}}
	obx_model_last_relation_id(model, {{.Model.LastRelationId.GetId}}, {{.Model.LastRelationId.GetUid}});
	{{- end}}
	return model; // NOTE: the returned model will contain error information if an error occurred.
}

#ifdef __cplusplus
}
#endif
`))
