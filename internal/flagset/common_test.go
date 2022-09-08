package flagset

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	"github.com/spf13/cobra"
)

type FlagTest struct {
	flag   string
	wanted string
}

func Test_GetFlagVarName(t *testing.T) {
	testCases := make([]FlagTest, 3)
	testCases[0] = FlagTest{"sample", "KUBEFIRST_SAMPLE"}
	testCases[1] = FlagTest{"sample-01", "KUBEFIRST_SAMPLE_01"}
	testCases[2] = FlagTest{"sample-01-ab", "KUBEFIRST_SAMPLE_01_AB"}
	for _, testCase := range testCases {
		producedValue := GetFlagVarName(testCase.flag)
		if producedValue != testCase.wanted {
			t.Errorf("GetFlagVarName was incorrect, got: %s, want: %s.", producedValue, testCase.wanted)
		}
	}

}

// based on: https://gianarb.it/blog/golang-mockmania-cli-command-with-cobra

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hugo",
		Short: "Hugo is a very fast static site generator",
		RunE: func(cmd *cobra.Command, args []string) error {
			flagValue, err := ReadConfigString(cmd, "sample")
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), err.Error())
			}
			fmt.Fprintf(cmd.OutOrStdout(), flagValue)
			return nil
		},
	}
	cmd.Flags().String("sample", "not-set", "This is a very important input.")
	return cmd
}

func Test_DefineSource_set_by_flag(t *testing.T) {
	cmd := NewRootCmd()
	b := bytes.NewBufferString("")
	os.Unsetenv("KUBEFIRST_SAMPLE")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"--sample", "set-by-flag"})
	cmd.Execute()
	out, err := ioutil.ReadAll(b)
	if err != nil {
		t.Error(err)
	}
	if string(out) != "set-by-flag" {
		t.Errorf("expected \"%s\" got \"%s\"", "set-by-flag", string(out))
	}
}

func Test_DefineSource_set_by_var(t *testing.T) {
	cmd := NewRootCmd()
	b := bytes.NewBufferString("")
	os.Setenv("KUBEFIRST_SAMPLE", "set-by-var")
	cmd.SetOut(b)
	cmd.Execute()
	out, err := ioutil.ReadAll(b)
	if err != nil {
		t.Error(err)
	}
	if string(out) != "set-by-var" {
		t.Errorf("expected \"%s\" got \"%s\"", "set-by-var", string(out))
	}
}

func Test_DefineSource_notSet(t *testing.T) {
	cmd := NewRootCmd()
	b := bytes.NewBufferString("")
	os.Unsetenv("KUBEFIRST_SAMPLE")
	cmd.SetOut(b)
	cmd.Execute()
	out, err := ioutil.ReadAll(b)
	if err != nil {
		t.Error(err)
	}
	if string(out) != "not-set" {
		t.Errorf("expected \"%s\" got \"%s\"", "not-set", string(out))
	}
}

func NewRootCmdBool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hugo",
		Short: "Hugo is a very fast static site generator",
		RunE: func(cmd *cobra.Command, args []string) error {
			flagValue, err := ReadConfigBool(cmd, "sample")
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), err.Error())
			}
			fmt.Fprintf(cmd.OutOrStdout(), strconv.FormatBool(flagValue))
			return nil
		},
	}
	cmd.Flags().Bool("sample", false, "This is a very important input.")
	return cmd
}

func Test_DefineSource_set_by_flag_bool(t *testing.T) {
	cmd := NewRootCmdBool()
	b := bytes.NewBufferString("")
	os.Setenv("KUBEFIRST_SAMPLE", "TRUE")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"--sample"})
	cmd.Execute()
	out, err := ioutil.ReadAll(b)
	if err != nil {
		t.Error(err)
	}
	if string(out) != "true" {
		t.Errorf("expected \"%s\" got \"%s\"", "true", string(out))
	}
}