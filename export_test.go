/*
Copyright © 2021  Red Hat, Inc.

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

package main

// Export for testing
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/ccx-notification-writer/packages/export_test.html
//
// This source file contains name aliases of all package-private functions
// that need to be called from unit tests. Aliases should start with uppercase
// letter because unit tests belong to different package.
//
// Please look into the following blogpost:
// https://medium.com/@robiplus/golang-trick-export-for-test-aa16cbd7b8cd
// to see why this trick is needed.
var (
	// functions from the logging.go source file
	LogDuration             = logDuration
	LogMessageInfo          = logMessageInfo
	LogMessageError         = logMessageError
	LogUnparsedMessageError = logUnparsedMessageError
	LogMessageWarning       = logMessageWarning

	// functions from the ccx_notification_writer.go source file
	ShowVersion         = showVersion
	ShowAuthors         = showAuthors
	ShowConfiguration   = showConfiguration
	DoSelectedOperation = doSelectedOperation

	// functions from consumer.go source file
	ParseMessage  = parseMessage
	ShrinkMessage = shrinkMessage

	// functions from storage.go source file
	DropTableStatement       = dropTableStatement
	DropIndexStatement       = dropIndexStatement
	DeleteFromTableStatement = deleteFromTableStatement
)
