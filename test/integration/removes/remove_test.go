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

package rename

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/objectbox/objectbox-generator/internal/generator"
	"github.com/objectbox/objectbox-generator/internal/generator/model"
	"github.com/objectbox/objectbox-generator/test/assert"
	"github.com/objectbox/objectbox-generator/test/integration"
)

func getUid(t *testing.T, id model.IdUid) model.Uid {
	uid, err := id.GetUid()
	assert.NoErr(t, err)
	return uid
}

func generateForStep(t *testing.T, conf *integration.CCppTestConf, step, schemaContents string) {
	conf.CreateCMake(t, integration.Cpp11, step+".cpp")
	var schema = filepath.Join(conf.Cmake.ConfDir, "schema.fbs")
	defer os.Remove(schema)
	assert.NoErr(t, ioutil.WriteFile(schema, []byte(schemaContents), 0600))
	conf.Generate(t, "", "")
}

func TestCpp(t *testing.T) {
	dbDir, err := ioutil.TempDir("", "generator-test-db")
	assert.NoErr(t, err)
	defer os.RemoveAll(dbDir)
	var envVars = []string{"dbDir=" + dbDir}

	conf := &integration.CCppTestConf{}
	defer conf.Cleanup()

	// STEP-1 start
	generateForStep(t, conf, "step-1", `
/// This entity will be removed in step 2
/// objectbox:relation(to=EntityB,name=standaloneRel)
table EntityA {
	id:uint64;

	/// objectbox:index
	name:string;

	/// objectbox:relation=EntityB
	relId:ulong;
}

table EntityB {
	id:uint64;
	name:string;
}`)

	modelJSONFile := generator.ModelInfoFile(conf.Cmake.ConfDir)
	modelInfo, err := model.LoadModelFromJSONFile(modelJSONFile)
	assert.NoErr(t, err)

	// collect UIDs that are expected to be retired in step 2
	entityUids := make([]model.Uid, 0)
	propertyUids := make([]model.Uid, 0)
	standaloneRelUids := make([]model.Uid, 0)
	indexUids := make([]model.Uid, 0)

	assert.Eq(t, 2, len(modelInfo.Entities))
	assert.Eq(t, "EntityA", modelInfo.Entities[0].Name)
	assert.Eq(t, "EntityB", modelInfo.Entities[1].Name)

	entityUids = append(entityUids, getUid(t, modelInfo.Entities[0].Id))

	assert.Eq(t, 3, len(modelInfo.Entities[0].Properties))
	propertyUids = append(propertyUids, getUid(t, modelInfo.Entities[0].Properties[0].Id))
	propertyUids = append(propertyUids, getUid(t, modelInfo.Entities[0].Properties[1].Id))
	propertyUids = append(propertyUids, getUid(t, modelInfo.Entities[0].Properties[2].Id))

	indexUids = append(indexUids, getUid(t, *modelInfo.Entities[0].Properties[1].IndexId))
	indexUids = append(indexUids, getUid(t, *modelInfo.Entities[0].Properties[2].IndexId))

	assert.Eq(t, 1, len(modelInfo.Entities[0].Relations))
	standaloneRelUids = append(standaloneRelUids, getUid(t, modelInfo.Entities[0].Relations[0].Id))

	conf.Build(t)
	conf.Run(t, envVars)
	// STEP-1 end

	// STEP-2 start
	generateForStep(t, conf, "step-2", `
table EntityB {
	id:uint64;
	name:string;
}`)

	modelInfo, err = model.LoadModelFromJSONFile(modelJSONFile)
	assert.NoErr(t, err)
	assert.Eq(t, 1, len(modelInfo.Entities))
	assert.Eq(t, "EntityB", modelInfo.Entities[0].Name)
	assert.EqItems(t, modelInfo.RetiredEntityUids, entityUids)
	assert.EqItems(t, modelInfo.RetiredIndexUids, indexUids)
	assert.EqItems(t, modelInfo.RetiredPropertyUids, propertyUids)
	assert.EqItems(t, modelInfo.RetiredRelationUids, standaloneRelUids)
	conf.Build(t)
	conf.Run(t, envVars)
	// STEP-2 end
}
