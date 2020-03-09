package check

import (
	"context"
	"io"
	"os"

	"github.com/rclone/rclone/cmd"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/config/flags"
	"github.com/rclone/rclone/fs/operations"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Globals
var (
	download = false
	oneway   = false
	combined = ""
	dstOnly  = ""
	srcOnly  = ""
	match    = ""
	differ   = ""
)

func init() {
	cmd.Root.AddCommand(commandDefinition)
	cmdFlags := commandDefinition.Flags()
	AddFlags(cmdFlags)
}

// AddFlags adds the check flags to the cmdFlags command
func AddFlags(cmdFlags *pflag.FlagSet) {
	flags.BoolVarP(cmdFlags, &oneway, "one-way", "", oneway, "Check one way only, source files must exist on remote")
	flags.StringVarP(cmdFlags, &combined, "combined", "", combined, "Make a combined report of changes to this file")
	flags.StringVarP(cmdFlags, &dstOnly, "dst-only", "", dstOnly, "Report all files missing from the source to this file")
	flags.StringVarP(cmdFlags, &srcOnly, "src-only", "", srcOnly, "Report all files missing from the destination to this file")
	flags.StringVarP(cmdFlags, &match, "match", "", match, "Report all matching files to this file")
	flags.StringVarP(cmdFlags, &differ, "differ", "", differ, "Report all non-matching files to this file")
}

// FlagsHelp describes the flags for the help
var FlagsHelp = `

If you supply the --one-way flag, it will only check that files in source
match the files in destination, not the other way around. Meaning extra files in
destination that are not in the source will not trigger an error.
`

// GetCheckOpt gets the options corresponding to the check flags
func GetCheckOpt(fsrc, fdst fs.Fs) (opt *operations.CheckOpt, close func(), err error) {
	closers := []io.Closer{}

	opt = &operations.CheckOpt{
		Fsrc:   fsrc,
		Fdst:   fdst,
		OneWay: oneway,
	}

	open := func(name string, pout *io.Writer) error {
		if name == "" {
			return nil
		}
		if name == "-" {
			*pout = os.Stdout
			return nil
		}
		out, err := os.Create(name)
		if err != nil {
			return err
		}
		*pout = out
		closers = append(closers, out)
		return nil
	}

	if err = open(combined, &opt.Combined); err != nil {
		return nil, nil, err
	}
	if err = open(dstOnly, &opt.DstOnly); err != nil {
		return nil, nil, err
	}
	if err = open(srcOnly, &opt.SrcOnly); err != nil {
		return nil, nil, err
	}
	if err = open(match, &opt.Match); err != nil {
		return nil, nil, err
	}
	if err = open(differ, &opt.Differ); err != nil {
		return nil, nil, err
	}

	close = func() {
		for _, closer := range closers {
			err := closer.Close()
			if err != nil {
				fs.Errorf(nil, "Failed to close report output: %v", err)
			}
		}
	}

	return opt, close, nil

}

var commandDefinition = &cobra.Command{
	Use:   "check source:path dest:path",
	Short: `Checks the files in the source and destination match.`,
	Long: `
Checks the files in the source and destination match.  It compares
sizes and hashes (MD5 or SHA1) and logs a report of files which don't
match.  It doesn't alter the source or destination.

If you supply the --size-only flag, it will only compare the sizes not
the hashes as well.  Use this for a quick check.

If you supply the --download flag, it will download the data from
both remotes and check them against each other on the fly.  This can
be useful for remotes that don't support hashes or if you really want
to check all the data.
` + FlagsHelp,
	Run: func(command *cobra.Command, args []string) {
		cmd.CheckArgs(2, 2, command, args)
		fsrc, fdst := cmd.NewFsSrcDst(args)
		cmd.Run(false, false, command, func() error {
			opt, close, err := GetCheckOpt(fsrc, fdst)
			if err != nil {
				return err
			}
			defer close()
			if download {
				return operations.CheckDownload(context.Background(), opt)
			}
			return operations.Check(context.Background(), opt)
		})
	},
}
