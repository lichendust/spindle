package main

import (
	"fmt"
	"time"
	"os/exec"
	"github.com/dop251/goja"
)

const script_wrapper = `function _() {%s};_()`

// future-proofing structure to for adding
// addition page APIs as needed
type script_markup struct {
	Vars map[string]string
}

type script_result struct {
	success       bool
	wants_cache   bool
	text   string
}

func call_script(vars map[string]string, program_text string, args []string) *script_result {
	vm := goja.New()

	// per-call setup
	vm.Set("args", args)
	vm.Set("the_page", &script_markup {vars})

	// configuration
	vm.Set("cache_return", false)

	// functions
	vm.Set("print", console_print)
	vm.Set("sprint", sprint)

	// returns unix time stamp of last git commit
	vm.Set("time", func(a ...string) string {
		// return passed file path or self if none
		if len(a) > 0 {
			return git_commit(a[0])
		}
		return git_commit(vars["raw_path"])
	})

	// execution
	v, err := vm.RunString(sprint(script_wrapper, program_text))

	result_data := script_result {}

	if err != nil {
		result_data.success = false
		result_data.text = err.Error()
		return &result_data
	}

	if v := v.Export(); v != nil {
		result_data.success = true
		result_data.text = v.(string)
	}

	if cache_return := vm.Get("cache_return"); cache_return.ToBoolean() == true {
		result_data.wants_cache = true
	}

	fmt.Printf("called for %q\n", result_data.text)

	return &result_data
}

// @experimental
// just returns unix time
func git_commit(path string) string {
	fmt.Println(path)

	cmd := exec.Command("git", "log", "-n", "1", "--format=%ct", "--", path)

	if result, err := cmd.Output(); err == nil {
		return string(result)
	}

	return fmt.Sprintf("%d", time.Now().Unix())
}