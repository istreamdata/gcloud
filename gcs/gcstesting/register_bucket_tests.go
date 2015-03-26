// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gcstesting

import (
	"reflect"
	"strings"

	"github.com/jacobsa/gcloud/gcs"
	"github.com/googlecloudplatform/gcsfuse/timeutil"
	"github.com/jacobsa/ogletest"
	"github.com/jacobsa/ogletest/srcutil"
)

// Dependencies needed for tests registered by RegisterBucketTests.
type BucketTestDeps struct {
	// An initialized, empty bucket.
	Bucket gcs.Bucket

	// A clock matching the bucket's notion of time.
	Clock timeutil.Clock
}

// An interface that all bucket tests must implement.
type bucketTestSetUpInterface interface {
	setUpBucketTest(deps BucketTestDeps)
}

func getSuiteName(suiteType reflect.Type) string {
	return strings.Title(suiteType.Name())
}

func isExported(name string) bool {
	return len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z'
}

func getTestMethods(suitePointerType reflect.Type) []reflect.Method {
	var exportedMethods []reflect.Method
	for _, m := range srcutil.GetMethodsInSourceOrder(suitePointerType) {
		if isExported(m.Name) {
			exportedMethods = append(exportedMethods, m)
		}
	}

	return exportedMethods
}

func registerTestSuite(
	makeDeps func() BucketTestDeps,
	prototype bucketTestSetUpInterface) {
	suitePointerType := reflect.TypeOf(prototype)
	suiteType := suitePointerType.Elem()

	// We don't need anything fancy at the suite level.
	var ts ogletest.TestSuite
	ts.Name = getSuiteName(suiteType)

	// For each method, we create a test function.
	for _, method := range getTestMethods(suitePointerType) {
		var tf ogletest.TestFunction
		tf.Name = method.Name

		// Create an instance to be shared among SetUp and the test function itself.
		var instance reflect.Value = reflect.New(suiteType)

		// SetUp should create a bucket and then initialize the suite object,
		// remembering that the suite implements bucketTestSetUpInterface.
		tf.SetUp = func(*ogletest.TestInfo) {
			deps := makeDeps()
			instance.Interface().(bucketTestSetUpInterface).setUpBucketTest(deps)
		}

		// The test function itself should simply invoke the method.
		methodCopy := method
		tf.Run = func() {
			methodCopy.Func.Call([]reflect.Value{instance})
		}

		// Save the test function.
		ts.TestFunctions = append(ts.TestFunctions, tf)
	}

	// Register the suite.
	ogletest.Register(ts)
}

// Given a function that returns appropriate test depencencies, register test
// suites that exercise the buckets returned by the function with ogletest.
func RegisterBucketTests(makeDeps func() BucketTestDeps) {
	// A list of empty instances of the test suites we want to register.
	suitePrototypes := []bucketTestSetUpInterface{
		&createTest{},
		&readTest{},
		&statTest{},
		&updateTest{},
		&deleteTest{},
		&listTest{},
	}

	// Register each.
	for _, suitePrototype := range suitePrototypes {
		registerTestSuite(makeDeps, suitePrototype)
	}
}