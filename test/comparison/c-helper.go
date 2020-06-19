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

package comparison

import (
	"io/ioutil"
	"path"
	"path/filepath"
	"testing"

	"github.com/objectbox/objectbox-generator/internal/generator"
	cgenerator "github.com/objectbox/objectbox-generator/internal/generator/c"
	"github.com/objectbox/objectbox-generator/test/assert"
	"github.com/objectbox/objectbox-generator/test/cmake"
)

type cTestHelper struct {
	cpp bool
}

func (h cTestHelper) generatorFor(t *testing.T, conf testSpec, sourceFile string, genDir string) generator.CodeGenerator {
	// make a copy of the default generator
	var gen = *conf.generator.(*cgenerator.CGenerator)
	gen.OutPath = genDir
	return &gen
}

func (cTestHelper) prepareTempDir(t *testing.T, conf testSpec, srcDir, tempDir, tempRoot string) func(err error) error {
	return nil
}

func (h cTestHelper) build(t *testing.T, conf testSpec, dir string, expectedError error, errorTransformer func(err error) error) {
	includeDir, err := filepath.Abs(dir) // main.c/cpp will include generated headers from here
	assert.NoErr(t, err)

	build := cmake.Cmake{
		Name:        "compilation-test",
		IsCpp:       h.cpp,
		IncludeDirs: []string{includeDir},
		LinkLibs:    []string{"objectbox"},
	}
	assert.NoErr(t, build.CreateTempDirs())
	defer build.RemoveTempDirs()

	var mainFile string
	if build.IsCpp {
		build.Standard = 11
		mainFile = path.Join(build.ConfDir, "main.cpp")
	} else {
		build.Standard = 99
		mainFile = path.Join(build.ConfDir, "main.c")
	}

	build.Files = append(build.Files, mainFile)

	assert.NoErr(t, build.WriteCMakeListsTxt())

	{ // write main.c/cpp to the conf dir - a simple one, just include all sources
		var mainSrc = ""

		files, err := ioutil.ReadDir(includeDir)
		assert.NoErr(t, err)
		for _, file := range files {
			if conf.generator.IsGeneratedFile(file.Name()) {
				mainSrc = mainSrc + "#include \"" + file.Name() + "\"\n"
			}
		}
		mainSrc = mainSrc + "int main(){ return 0; }\n\n"
		assert.NoErr(t, ioutil.WriteFile(mainFile, []byte(mainSrc), 0600))
	}

	if stdOut, stdErr, err := build.Configure(); err != nil {
		assert.Failf(t, "cmake configuration failed: \n%s\n%s\n%s", stdOut, stdErr, err)
	} else {
		t.Logf("configuration output:\n%s", string(stdOut))
	}

	if stdOut, stdErr, err := build.Build(); err != nil {
		checkBuildError(t, errorTransformer, stdOut, stdErr, err, expectedError)
		assert.Failf(t, "cmake build failed: \n%s\n%s\n%s", stdOut, stdErr, err)
	} else {
		t.Logf("build output:\n%s", string(stdOut))
	}
}
