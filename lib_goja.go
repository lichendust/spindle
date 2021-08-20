package main

import (
	"time"
	"github.com/dop251/goja"
)

const script_wrapper = `function _() {%s};_()`

func new_script_vm() *goja.Runtime {
	vm := goja.New()

	vm.SetFieldNameMapper(goja.UncapFieldNameMapper())

	// builtins
	vm.Set("print", console_handler.print)
	vm.Set("sprint", sprint)

	return vm
}

// future-proofing structure to for adding
// addition page APIs as needed
type script_markup struct {
	Vars map[string]string
}

func call_script(vars map[string]string, program_text string, args []string) (string, bool) {
	vm := new_script_vm()

	// per-call setup
	vm.Set("args", args)
	vm.Set("the_page", &script_markup {vars})

	// execution
	v, err := vm.RunString(sprint(script_wrapper, program_text))

	if err != nil {
		return err.Error(), false
	}

	if v := v.Export(); v != nil {
		return v.(string), true
	}

	return "", true
}