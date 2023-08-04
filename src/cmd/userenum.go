package cmd

import (
	"errors"
	"regexp"
	"strings"

	"TeamsUserEnum/src/teams"

	"github.com/spf13/cobra"
)

var emailFile string
var email string
var token string
var threads int
var outputFile string

// userenumCmd represents the userenum command
var userenumCmd = &cobra.Command{
	Use:   "userenum",                            // Name of the command
	Short: "User enumeration on Microsoft Teams", // Short description
	Long: `Users can be enumerated on Microsoft Teams with the search features.
This tool validates an email address or a list of email addresses.
If these emails exist the presence of the user is retrieved as well as the device used to connect`,

	// Check if the arguments are valid
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if emailFile == "" && email == "" {
			return errors.New("argument -f or -e required")
		} else if emailFile != "" && email != "" {
			return errors.New("only argument -f or -e should be specified")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if strings.Contains(token, "&") {
			token = strings.ReplaceAll(token, "&", "%26")
		}
		re := regexp.MustCompile(`ey.*?%`)                    // Extract the base64 part of the token
		match := re.FindStringSubmatch(token)                 // Find the needed portion
		token = "Bearer " + strings.TrimSuffix(match[0], "%") // set token to the needed format

		if email != "" {
			teams.Enumuser(email, token, verbose) // Enumerate a single email
		} else {
			teams.Parsefile(emailFile, token, verbose, threads, outputFile) // Enumerate a list of emails
		}
	},
}

func init() {
	rootCmd.AddCommand(userenumCmd) // Add the command to the root command

	userenumCmd.Flags().StringVarP(&emailFile, "file", "f", "", "File containing the email address")
	userenumCmd.Flags().StringVarP(&email, "email", "e", "", "Email address")
	userenumCmd.Flags().StringVarP(&token, "token", "t", "", "Full Bearer token")
	userenumCmd.Flags().IntVarP(&threads, "threads", "T", 5, "Number of threads to use")
	userenumCmd.Flags().StringVarP(&outputFile, "Output File", "o", "", "File to write found emails to")
	userenumCmd.MarkFlagRequired("token")
}
