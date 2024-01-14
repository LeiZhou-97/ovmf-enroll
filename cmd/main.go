package main

import (
	"fmt"
	"ovmf/ovmf-enroll/pkg/varenroll"

	"github.com/spf13/cobra"
)

var (
	ovmf = &varenroll.OVMFctl{}
	rootCmd = &cobra.Command{
		Use:   "ovmf-enroll-tool",
		Short: "A tool of enrolling variable into OVMF",
	}
)

func main() {
	rootCmd.Execute()

	err := varenroll.VarEnroll(*ovmf)
	if err != nil {
		fmt.Println(err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&ovmf.InputFile, "file", "f", "", "OVMF input file")
	rootCmd.PersistentFlags().StringVarP(&ovmf.OutputFile, "output", "o", "", "OVMF output file after the var is enrolled")
	rootCmd.PersistentFlags().StringVarP(&ovmf.Name, "name", "n", "", "Name of the variable to be enrolled, such as PK/KEK/db/dbx/SecureBootEnable etc")
	rootCmd.PersistentFlags().StringVarP(&ovmf.Guid, "guid", "g", "", "For PK/KEK/db/dbx,it's guid of signature owner. For other variable it's vendor guid")
	rootCmd.PersistentFlags().StringVarP(&ovmf.Attributes, "attributes", "a", "", "For PK/KEK/db/dbx, ignored. For other variables means its attribute, e.g 0x3")
	rootCmd.PersistentFlags().StringVarP(&ovmf.DataFile, "data", "d", "", "For PK/KEK/db/dbx, it's the cert file. Otherwise it's the payload of the variables.")

}
