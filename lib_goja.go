package main

import (
	"time"
	"os/exec"
	"strconv"
	"strings"
	"path/filepath"
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

	vm.SetFieldNameMapper(goja.UncapFieldNameMapper())

	// per-call setup
	vm.Set("args", args)
	vm.Set("the_page", &script_markup {vars})

	vm.Set("get_page", func(path string) *script_markup {
		path = filepath.Join("source", path)

		if is_dir(path) {
			path = filepath.Join(path, "index.x")
		} else {
			path += ".x"
		}

		page, ok := load_page(path)

		if !ok {
			console_print("bad times with that page")
			return nil
		}

		return &script_markup {page.vars}
	})

	// configuration
	vm.Set("cache_return", false)

	// functions
	vm.Set("print",      console_print)
	vm.Set("sprint",     sprint)
	vm.Set("title_case", make_title)

	// vm.Set("import", )

	// returns unix time stamp of last git commit
	vm.Set("git_time", func(format_string string) string {
		return git_commit(vars["raw_path"], format_string)
	})

	// execution
	v, err := vm.RunString(sprint(script_wrapper, program_text))

	result_data := script_result {}

	if err != nil {
		result_data.success = false
		result_data.text = err.Error()
		return &result_data
	}

	result_data.success = true

	if v := v.Export(); v != nil {
		result_data.text = v.(string)
	}

	if cache_return := vm.Get("cache_return"); cache_return.ToBoolean() == true {
		result_data.wants_cache = true
	}

	return &result_data
}

// @experimental
// just returns unix time
func git_commit(path, format_string string) string {
	cmd := exec.Command("git", "log", "-n", "1", "--format=%ct", "--", path)

	result, err := cmd.Output()

	if err != nil {
		return ""
	}

	i, err := strconv.ParseInt(strings.TrimSpace(string(result)), 10, 64)

	if err != nil {
		return ""
	}

	the_time := time.Unix(i, 0)

	return nsdate(the_time, format_string)
}