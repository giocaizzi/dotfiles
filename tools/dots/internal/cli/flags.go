package cli

import "flag"

func newFlags(name string) *flag.FlagSet { return flag.NewFlagSet(name, flag.ContinueOnError) }
