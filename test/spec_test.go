// Copyright (c) 2025 Mike Schinkel
// Portions Copyright (c) 2014 Alex Kalyvitis

package mustache_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	//"runtime/debug"
	"strings"
	"testing"
	"text/tabwriter"

	"github.com/alexkappa/mustache"
)

const (
	specDir    = "./spec"
	specRepo   = "https://github.com/mustache/spec.git"
	tmpSpecDir = "/tmp/mustache-spec" // Stable directory name
)

type SpecTest struct {
	Name     string            `json:"name"`
	Data     interface{}       `json:"data"`
	Expected string            `json:"expected"`
	Template string            `json:"template"`
	Partials map[string]string `json:"partials"`
	Desc     string            `json:"desc"`
}

type Spec struct {
	Filename string     `json:"-"`
	Overview string     `json:"overview"`
	Tests    []SpecTest `json:"tests"`
}

var specs = make(map[string]Spec)

func TestSpec(t *testing.T) {
	err := ensureSpecs(t)
	if err != nil {
		t.Log("Spec conformation tests are not available.")
		t.Logf("Unable to git clone or pull from %s", specRepo)
		env := os.Getenv("SKIP_SPECS")
		if env == "" {
			t.Error("Set SKIP_SPECS=yes to bypass spec conformance tests.")
			return
		}
		t.Skipf("Skipping spec confirmation tests because SKIP_SPECS=%s", env)
	}
	for _, spec := range specs {
		t.Run(filepath.Base(spec.Filename), testSpecFunc(t, spec, -1))
	}
}

var write = MustFprintf

func testSpecFunc(t *testing.T, s Spec, testNum int) func(t *testing.T) {
	t.Helper()
	return func(t *testing.T) {

		t.Parallel() // run each spec test concurrently.

		// Create a new writer to write information about the test, and in
		// the event the test fails we print it to the screen.
		w := newTestWriter()

		MustFprintf(w, "%s\n", s.Filename)
		for num, tt := range s.Tests {
			t.Run(tt.Name, func(t *testing.T) {
				if num != -1 && testNum != -1 && num != testNum {
					// Don't run a test if we haven't requested to run them all (testNum==-1) and if
					// we are not on the requested test num.
					return
				}
				// If num == -1 or num == test num, run the test

				// Handle and recover from panics.
				//goland:noinspection GoDeferInLoop
				defer func() {
					if r := recover(); r != nil {
						var calls string
						reason := fmt.Sprintf("%v", r)
						if strings.HasPrefix(reason, "runtime error:") {
							calls = string(debug.Stack())
						} else {
							calls = errorTriggeredBy()
						}
						write(w, "%s", "\n")
						write(w, "Name:     %s\n", tt.Name)
						write(w, "Testing:  %s\n", tt.Desc)
						write(w, "Template: %q\n", tt.Template)
						write(w, "Data:     %q\n", tt.Data)
						write(w, "Error:    %s\n", reason)
						write(w, "Calls:\n%s\n", calls)
						t.Fatal(w.String())
					}
				}()

				//goland:noinspection GoDeferInLoop
				defer w.Reset() // Reset writer after each test.

				// Parse the template and report errors.
				template := mustache.New()
				if err := template.ParseString(tt.Template); err != nil {
					MustFprintf(w, "Error: %s\n", err)
					t.Fatal(w.String())
				}

				// If partials were present in the spec test iterate and test each
				// one.
				for n, s := range tt.Partials {
					p := mustache.New(Name(n))
					if err := p.ParseString(s); err != nil {

						MustFprintf(w, "Partial : %s> %q\n", n, s)
						MustFprintf(w, "Error: %s\n", err)
						t.Fatal(w.String())
					}
					template.Option(Partial(p))
				}

				output, err := template.RenderString(tt.Data)
				if err != nil {
					MustFprintf(w, "Error: %s\n", err)
					t.Fatal(w.String())
				}

				if output != tt.Expected {
					write(w, "\n")
					write(w, "Name:     %s\n", tt.Name)
					write(w, "Tree:     %+v\n", template.Elems)
					write(w, "Data:     %q\n", tt.Data)
					write(w, "Template: %q\n", tt.Template)
					write(w, "Partials: %q\n", tt.Partials)
					write(w, "Expected: %q\n", tt.Expected)
					write(w, "Have:     %q\n", output)
					t.Fatal(w.String())
				}
			})
		}
	}
}

func ensureSpecs(t *testing.T) (err error) {
	var files []os.DirEntry
	t.Helper()

	errs := make([]error, 0)
	err = updateSpecs(t)

	if err != nil {
		err = fmt.Errorf("failed to update specs: %w", err)
		goto end
	}

	files, err = os.ReadDir(specDir)
	if err != nil {
		err = fmt.Errorf("failed to read specs directory %s: %w", specDir, err)
		goto end
	}

	for _, file := range files {
		fileName := file.Name()
		if !strings.HasSuffix(fileName, ".json") {
			continue
		}
		//if fileName[0] == '~' {
		//	// Bypass the optional specs, for now
		//	continue
		//}
		var spec Spec
		var f *os.File
		filename := filepath.Join(specDir, file.Name())
		f, err = os.Open(filename)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to read specs file %s: %w", filename, err))
			continue
		}
		err = json.NewDecoder(f).Decode(&spec)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to decode JSON from specs file %s: %w", filename, err))
			continue
		}
		spec.Filename = filename
		specs[strings.TrimSuffix(file.Name(), ".json")] = spec
	}
	if len(errs) != 0 {
		err = errors.Join(errs...)
	}
end:
	return err
}

type testWriter struct {
	b *bytes.Buffer
	w *tabwriter.Writer
}

func newTestWriter() *testWriter {
	t := &testWriter{}
	t.b = bytes.NewBuffer(nil)
	t.w = tabwriter.NewWriter(t.b, 0, 8, 0, ':', 0)
	return t
}

func (t *testWriter) Write(buf []byte) (n int, err error) {
	return t.w.Write(buf)
}
func (t *testWriter) String() string {
	must("flush", t.w.Flush())
	return t.b.String()
}
func (t *testWriter) Reset() {
	must("flush", t.w.Flush())
	t.b.Reset()
}

var (
	Name        = mustache.Name
	Partial     = mustache.Partial
	MustFprintf = mustache.MustFprintf
)

func getLatestTag(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "tag", "--sort=-v:refname")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get tags: %w", err)
	}
	tags := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(tags) == 0 {
		return "", fmt.Errorf("no tags found in repository")
	}
	return tags[0], nil
}

func updateSpecs(t *testing.T) error {
	// Check if git is available
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git is required but not found in PATH: %w", err)
	}

	// Check if the repo already exists
	if _, err := os.Stat(tmpSpecDir); os.IsNotExist(err) {
		// Clone the repository if it doesn't exist
		cmd := exec.Command("git", "clone", specRepo, tmpSpecDir)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to clone spec repository: %s: %w", string(out), err)
		}
	} else {
		t.Logf("Pulling latest changes from Mustache spec repository...")
		// First ensure we are on the main branch
		cmd := exec.Command("git", "-C", tmpSpecDir, "checkout", "master")
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to checkout master branch: %s: %w", string(out), err)
		}
		// Pull latest changes
		cmd = exec.Command("git", "-C", tmpSpecDir, "pull")
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to pull latest changes: %s: %w", string(out), err)
		}
	}
	// Get latest tag
	tag, err := getLatestTag(tmpSpecDir)
	if err != nil {
		return fmt.Errorf("failed to get latest tag: %w", err)
	}

	// Checkout the latest tag
	cmd := exec.Command("git", "-C", tmpSpecDir, "checkout", tag)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to checkout tag %s: %s: %w", tag, string(out), err)
	}

	// Ensure our specs directory exists
	if err := os.MkdirAll(specDir, 0755); err != nil {
		return fmt.Errorf("failed to create spec directory: %w", err)
	}

	// Remove existing JSON files
	files, err := os.ReadDir(specDir)
	if err != nil {
		return fmt.Errorf("failed to read spec directory: %w", err)
	}
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		if err := os.Remove(filepath.Join(specDir, file.Name())); err != nil {
			return fmt.Errorf("failed to remove old spec file %s: %w", file.Name(), err)
		}
	}

	// Copy new JSON files
	sourceDir := filepath.Join(tmpSpecDir, "specs")
	files, err = os.ReadDir(sourceDir)
	if err != nil {
		return fmt.Errorf("failed to read source spec directory: %w", err)
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			src := filepath.Join(sourceDir, file.Name())
			dst := filepath.Join(specDir, file.Name())

			srcFile, err := os.Open(src)
			if err != nil {
				return fmt.Errorf("failed to open source file %s: %w", file.Name(), err)
			}

			//goland:noinspection GoDeferInLoop
			defer mustClose(t, srcFile, src)

			dstFile, err := os.Create(dst)
			if err != nil {
				return fmt.Errorf("failed to create destination file %s: %w", file.Name(), err)
			}

			//goland:noinspection GoDeferInLoop
			defer mustClose(t, dstFile, dst)

			if _, err = io.Copy(dstFile, srcFile); err != nil {
				return fmt.Errorf("failed to copy file %s: %w", file.Name(), err)
			}
		}
	}

	return nil
}

func mustClose(t *testing.T, c io.Closer, file string) {
	err := c.Close()
	if err != nil {
		t.Errorf("failed to close %s: %v", file, err)
	}
}

func errorTriggeredBy() (by string) {
	var call string
	var lines, calls []string

	buf := make([]byte, 4096)
	for {
		n := runtime.Stack(buf, false)
		if n >= len(buf) {
			buf = make([]byte, 2*len(buf))
			continue
		}
		skipNext := false
		for line := range strings.Lines(string(buf[:n])) {
			if skipNext {
				skipNext = false
				continue
			}
			if strings.HasPrefix(line, "goroutine") {
				continue
			}
			if strings.HasPrefix(line, "github.com/") {
				call = line
				continue
			}
			if strings.HasPrefix(line, "\t/") {
				if strings.HasSuffix(call, "errorTriggeredBy()\n") {
					continue
				}
				if strings.Contains(line, "/go/src/") {
					skipNext = true
					continue
				}
				if strings.Contains(line, "_test.go:") {
					calls = append(calls, call)
					lines = append(lines, line)
					continue
				}
				calls = append(calls, call)
				lines = append(lines, line)
				if len(calls) >= 2 {
					goto end
				}
			}
		}
		goto end
	}
end:
	spacer := strings.Repeat(" ", 11)
	for i, c := range calls {
		by = fmt.Sprintf("%s%s%s%s%s", by, spacer, c, spacer, lines[i])
	}
	return by
}
