// Package tasks provides a framework to asynchronously run cancellable tasks with simplified progress tracking.
//
// The most important type of task is the pipeline, which allows the composition of individual tasks, whose inputs
// and outputs can be connected using edges.
//
// This system is mostly designed for the GUI.
//
// My(darwin) justification for this complicated approach is that the usual pipeline composition patterns make progress
// tracking, cancellation and error handling quite manual and error-prone. I think this approach strikes a fair middle-ground
// between ultra-manual code and drifting too deep into DSL/embedded langauge territory.
package tasks
