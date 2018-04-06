package cobra

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/spf13/pflag"
)

// Annotations for Bash completion.
const (
	BashCompFilenameExt     = "cobra_annotation_bash_completion_***REMOVED***lename_extensions"
	BashCompCustom          = "cobra_annotation_bash_completion_custom"
	BashCompOneRequiredFlag = "cobra_annotation_bash_completion_one_required_flag"
	BashCompSubdirsInDir    = "cobra_annotation_bash_completion_subdirs_in_dir"
)

func writePreamble(buf *bytes.Buffer, name string) {
	buf.WriteString(fmt.Sprintf("# bash completion for %-36s -*- shell-script -*-\n", name))
	buf.WriteString(fmt.Sprintf(`
__%[1]s_debug()
{
    if [[ -n ${BASH_COMP_DEBUG_FILE} ]]; then
        echo "$*" >> "${BASH_COMP_DEBUG_FILE}"
    ***REMOVED***
}

# Homebrew on Macs have version 1.3 of bash-completion which doesn't include
# _init_completion. This is a very minimal version of that function.
__%[1]s_init_completion()
{
    COMPREPLY=()
    _get_comp_words_by_ref "$@" cur prev words cword
}

__%[1]s_index_of_word()
{
    local w word=$1
    shift
    index=0
    for w in "$@"; do
        [[ $w = "$word" ]] && return
        index=$((index+1))
    done
    index=-1
}

__%[1]s_contains_word()
{
    local w word=$1; shift
    for w in "$@"; do
        [[ $w = "$word" ]] && return
    done
    return 1
}

__%[1]s_handle_reply()
{
    __%[1]s_debug "${FUNCNAME[0]}"
    case $cur in
        -*)
            if [[ $(type -t compopt) = "builtin" ]]; then
                compopt -o nospace
            ***REMOVED***
            local allflags
            if [ ${#must_have_one_flag[@]} -ne 0 ]; then
                allflags=("${must_have_one_flag[@]}")
            ***REMOVED***
                allflags=("${flags[*]} ${two_word_flags[*]}")
            ***REMOVED***
            COMPREPLY=( $(compgen -W "${allflags[*]}" -- "$cur") )
            if [[ $(type -t compopt) = "builtin" ]]; then
                [[ "${COMPREPLY[0]}" == *= ]] || compopt +o nospace
            ***REMOVED***

            # complete after --flag=abc
            if [[ $cur == *=* ]]; then
                if [[ $(type -t compopt) = "builtin" ]]; then
                    compopt +o nospace
                ***REMOVED***

                local index flag
                flag="${cur%%=*}"
                __%[1]s_index_of_word "${flag}" "${flags_with_completion[@]}"
                COMPREPLY=()
                if [[ ${index} -ge 0 ]]; then
                    PREFIX=""
                    cur="${cur#*=}"
                    ${flags_completion[${index}]}
                    if [ -n "${ZSH_VERSION}" ]; then
                        # zsh completion needs --flag= pre***REMOVED***x
                        eval "COMPREPLY=( \"\${COMPREPLY[@]/#/${flag}=}\" )"
                    ***REMOVED***
                ***REMOVED***
            ***REMOVED***
            return 0;
            ;;
    esac

    # check if we are handling a flag with special work handling
    local index
    __%[1]s_index_of_word "${prev}" "${flags_with_completion[@]}"
    if [[ ${index} -ge 0 ]]; then
        ${flags_completion[${index}]}
        return
    ***REMOVED***

    # we are parsing a flag and don't have a special handler, no completion
    if [[ ${cur} != "${words[cword]}" ]]; then
        return
    ***REMOVED***

    local completions
    completions=("${commands[@]}")
    if [[ ${#must_have_one_noun[@]} -ne 0 ]]; then
        completions=("${must_have_one_noun[@]}")
    ***REMOVED***
    if [[ ${#must_have_one_flag[@]} -ne 0 ]]; then
        completions+=("${must_have_one_flag[@]}")
    ***REMOVED***
    COMPREPLY=( $(compgen -W "${completions[*]}" -- "$cur") )

    if [[ ${#COMPREPLY[@]} -eq 0 && ${#noun_aliases[@]} -gt 0 && ${#must_have_one_noun[@]} -ne 0 ]]; then
        COMPREPLY=( $(compgen -W "${noun_aliases[*]}" -- "$cur") )
    ***REMOVED***

    if [[ ${#COMPREPLY[@]} -eq 0 ]]; then
        declare -F __custom_func >/dev/null && __custom_func
    ***REMOVED***

    # available in bash-completion >= 2, not always present on macOS
    if declare -F __ltrim_colon_completions >/dev/null; then
        __ltrim_colon_completions "$cur"
    ***REMOVED***

    # If there is only 1 completion and it is a flag with an = it will be completed
    # but we don't want a space after the =
    if [[ "${#COMPREPLY[@]}" -eq "1" ]] && [[ $(type -t compopt) = "builtin" ]] && [[ "${COMPREPLY[0]}" == --*= ]]; then
       compopt -o nospace
    ***REMOVED***
}

# The arguments should be in the form "ext1|ext2|extn"
__%[1]s_handle_***REMOVED***lename_extension_flag()
{
    local ext="$1"
    _***REMOVED***ledir "@(${ext})"
}

__%[1]s_handle_subdirs_in_dir_flag()
{
    local dir="$1"
    pushd "${dir}" >/dev/null 2>&1 && _***REMOVED***ledir -d && popd >/dev/null 2>&1
}

__%[1]s_handle_flag()
{
    __%[1]s_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    # if a command required a flag, and we found it, unset must_have_one_flag()
    local flagname=${words[c]}
    local flagvalue
    # if the word contained an =
    if [[ ${words[c]} == *"="* ]]; then
        flagvalue=${flagname#*=} # take in as flagvalue after the =
        flagname=${flagname%%=*} # strip everything after the =
        flagname="${flagname}=" # but put the = back
    ***REMOVED***
    __%[1]s_debug "${FUNCNAME[0]}: looking for ${flagname}"
    if __%[1]s_contains_word "${flagname}" "${must_have_one_flag[@]}"; then
        must_have_one_flag=()
    ***REMOVED***

    # if you set a flag which only applies to this command, don't show subcommands
    if __%[1]s_contains_word "${flagname}" "${local_nonpersistent_flags[@]}"; then
      commands=()
    ***REMOVED***

    # keep flag value with flagname as flaghash
    # flaghash variable is an associative array which is only supported in bash > 3.
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        if [ -n "${flagvalue}" ] ; then
            flaghash[${flagname}]=${flagvalue}
        elif [ -n "${words[ $((c+1)) ]}" ] ; then
            flaghash[${flagname}]=${words[ $((c+1)) ]}
        ***REMOVED***
            flaghash[${flagname}]="true" # pad "true" for bool flag
        ***REMOVED***
    ***REMOVED***

    # skip the argument to a two word flag
    if __%[1]s_contains_word "${words[c]}" "${two_word_flags[@]}"; then
        c=$((c+1))
        # if we are looking for a flags value, don't show commands
        if [[ $c -eq $cword ]]; then
            commands=()
        ***REMOVED***
    ***REMOVED***

    c=$((c+1))

}

__%[1]s_handle_noun()
{
    __%[1]s_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    if __%[1]s_contains_word "${words[c]}" "${must_have_one_noun[@]}"; then
        must_have_one_noun=()
    elif __%[1]s_contains_word "${words[c]}" "${noun_aliases[@]}"; then
        must_have_one_noun=()
    ***REMOVED***

    nouns+=("${words[c]}")
    c=$((c+1))
}

__%[1]s_handle_command()
{
    __%[1]s_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    local next_command
    if [[ -n ${last_command} ]]; then
        next_command="_${last_command}_${words[c]//:/__}"
    ***REMOVED***
        if [[ $c -eq 0 ]]; then
            next_command="_%[1]s_root_command"
        ***REMOVED***
            next_command="_${words[c]//:/__}"
        ***REMOVED***
    ***REMOVED***
    c=$((c+1))
    __%[1]s_debug "${FUNCNAME[0]}: looking for ${next_command}"
    declare -F "$next_command" >/dev/null && $next_command
}

__%[1]s_handle_word()
{
    if [[ $c -ge $cword ]]; then
        __%[1]s_handle_reply
        return
    ***REMOVED***
    __%[1]s_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"
    if [[ "${words[c]}" == -* ]]; then
        __%[1]s_handle_flag
    elif __%[1]s_contains_word "${words[c]}" "${commands[@]}"; then
        __%[1]s_handle_command
    elif [[ $c -eq 0 ]]; then
        __%[1]s_handle_command
    ***REMOVED***
        __%[1]s_handle_noun
    ***REMOVED***
    __%[1]s_handle_word
}

`, name))
}

func writePostscript(buf *bytes.Buffer, name string) {
	name = strings.Replace(name, ":", "__", -1)
	buf.WriteString(fmt.Sprintf("__start_%s()\n", name))
	buf.WriteString(fmt.Sprintf(`{
    local cur prev words cword
    declare -A flaghash 2>/dev/null || :
    if declare -F _init_completion >/dev/null 2>&1; then
        _init_completion -s || return
    ***REMOVED***
        __%[1]s_init_completion -n "=" || return
    ***REMOVED***

    local c=0
    local flags=()
    local two_word_flags=()
    local local_nonpersistent_flags=()
    local flags_with_completion=()
    local flags_completion=()
    local commands=("%[1]s")
    local must_have_one_flag=()
    local must_have_one_noun=()
    local last_command
    local nouns=()

    __%[1]s_handle_word
}

`, name))
	buf.WriteString(fmt.Sprintf(`if [[ $(type -t compopt) = "builtin" ]]; then
    complete -o default -F __start_%s %s
***REMOVED***
    complete -o default -o nospace -F __start_%s %s
***REMOVED***

`, name, name, name, name))
	buf.WriteString("# ex: ts=4 sw=4 et ***REMOVED***letype=sh\n")
}

func writeCommands(buf *bytes.Buffer, cmd *Command) {
	buf.WriteString("    commands=()\n")
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c == cmd.helpCommand {
			continue
		}
		buf.WriteString(fmt.Sprintf("    commands+=(%q)\n", c.Name()))
	}
	buf.WriteString("\n")
}

func writeFlagHandler(buf *bytes.Buffer, name string, annotations map[string][]string, cmd *Command) {
	for key, value := range annotations {
		switch key {
		case BashCompFilenameExt:
			buf.WriteString(fmt.Sprintf("    flags_with_completion+=(%q)\n", name))

			var ext string
			if len(value) > 0 {
				ext = fmt.Sprintf("__%s_handle_***REMOVED***lename_extension_flag ", cmd.Root().Name()) + strings.Join(value, "|")
			} ***REMOVED*** {
				ext = "_***REMOVED***ledir"
			}
			buf.WriteString(fmt.Sprintf("    flags_completion+=(%q)\n", ext))
		case BashCompCustom:
			buf.WriteString(fmt.Sprintf("    flags_with_completion+=(%q)\n", name))
			if len(value) > 0 {
				handlers := strings.Join(value, "; ")
				buf.WriteString(fmt.Sprintf("    flags_completion+=(%q)\n", handlers))
			} ***REMOVED*** {
				buf.WriteString("    flags_completion+=(:)\n")
			}
		case BashCompSubdirsInDir:
			buf.WriteString(fmt.Sprintf("    flags_with_completion+=(%q)\n", name))

			var ext string
			if len(value) == 1 {
				ext = fmt.Sprintf("__%s_handle_subdirs_in_dir_flag ", cmd.Root().Name()) + value[0]
			} ***REMOVED*** {
				ext = "_***REMOVED***ledir -d"
			}
			buf.WriteString(fmt.Sprintf("    flags_completion+=(%q)\n", ext))
		}
	}
}

func writeShortFlag(buf *bytes.Buffer, flag *pflag.Flag, cmd *Command) {
	name := flag.Shorthand
	format := "    "
	if len(flag.NoOptDefVal) == 0 {
		format += "two_word_"
	}
	format += "flags+=(\"-%s\")\n"
	buf.WriteString(fmt.Sprintf(format, name))
	writeFlagHandler(buf, "-"+name, flag.Annotations, cmd)
}

func writeFlag(buf *bytes.Buffer, flag *pflag.Flag, cmd *Command) {
	name := flag.Name
	format := "    flags+=(\"--%s"
	if len(flag.NoOptDefVal) == 0 {
		format += "="
	}
	format += "\")\n"
	buf.WriteString(fmt.Sprintf(format, name))
	writeFlagHandler(buf, "--"+name, flag.Annotations, cmd)
}

func writeLocalNonPersistentFlag(buf *bytes.Buffer, flag *pflag.Flag) {
	name := flag.Name
	format := "    local_nonpersistent_flags+=(\"--%s"
	if len(flag.NoOptDefVal) == 0 {
		format += "="
	}
	format += "\")\n"
	buf.WriteString(fmt.Sprintf(format, name))
}

func writeFlags(buf *bytes.Buffer, cmd *Command) {
	buf.WriteString(`    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

`)
	localNonPersistentFlags := cmd.LocalNonPersistentFlags()
	cmd.NonInheritedFlags().VisitAll(func(flag *pflag.Flag) {
		if nonCompletableFlag(flag) {
			return
		}
		writeFlag(buf, flag, cmd)
		if len(flag.Shorthand) > 0 {
			writeShortFlag(buf, flag, cmd)
		}
		if localNonPersistentFlags.Lookup(flag.Name) != nil {
			writeLocalNonPersistentFlag(buf, flag)
		}
	})
	cmd.InheritedFlags().VisitAll(func(flag *pflag.Flag) {
		if nonCompletableFlag(flag) {
			return
		}
		writeFlag(buf, flag, cmd)
		if len(flag.Shorthand) > 0 {
			writeShortFlag(buf, flag, cmd)
		}
	})

	buf.WriteString("\n")
}

func writeRequiredFlag(buf *bytes.Buffer, cmd *Command) {
	buf.WriteString("    must_have_one_flag=()\n")
	flags := cmd.NonInheritedFlags()
	flags.VisitAll(func(flag *pflag.Flag) {
		if nonCompletableFlag(flag) {
			return
		}
		for key := range flag.Annotations {
			switch key {
			case BashCompOneRequiredFlag:
				format := "    must_have_one_flag+=(\"--%s"
				if flag.Value.Type() != "bool" {
					format += "="
				}
				format += "\")\n"
				buf.WriteString(fmt.Sprintf(format, flag.Name))

				if len(flag.Shorthand) > 0 {
					buf.WriteString(fmt.Sprintf("    must_have_one_flag+=(\"-%s\")\n", flag.Shorthand))
				}
			}
		}
	})
}

func writeRequiredNouns(buf *bytes.Buffer, cmd *Command) {
	buf.WriteString("    must_have_one_noun=()\n")
	sort.Sort(sort.StringSlice(cmd.ValidArgs))
	for _, value := range cmd.ValidArgs {
		buf.WriteString(fmt.Sprintf("    must_have_one_noun+=(%q)\n", value))
	}
}

func writeArgAliases(buf *bytes.Buffer, cmd *Command) {
	buf.WriteString("    noun_aliases=()\n")
	sort.Sort(sort.StringSlice(cmd.ArgAliases))
	for _, value := range cmd.ArgAliases {
		buf.WriteString(fmt.Sprintf("    noun_aliases+=(%q)\n", value))
	}
}

func gen(buf *bytes.Buffer, cmd *Command) {
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c == cmd.helpCommand {
			continue
		}
		gen(buf, c)
	}
	commandName := cmd.CommandPath()
	commandName = strings.Replace(commandName, " ", "_", -1)
	commandName = strings.Replace(commandName, ":", "__", -1)

	if cmd.Root() == cmd {
		buf.WriteString(fmt.Sprintf("_%s_root_command()\n{\n", commandName))
	} ***REMOVED*** {
		buf.WriteString(fmt.Sprintf("_%s()\n{\n", commandName))
	}

	buf.WriteString(fmt.Sprintf("    last_command=%q\n", commandName))
	writeCommands(buf, cmd)
	writeFlags(buf, cmd)
	writeRequiredFlag(buf, cmd)
	writeRequiredNouns(buf, cmd)
	writeArgAliases(buf, cmd)
	buf.WriteString("}\n\n")
}

// GenBashCompletion generates bash completion ***REMOVED***le and writes to the passed writer.
func (c *Command) GenBashCompletion(w io.Writer) error {
	buf := new(bytes.Buffer)
	writePreamble(buf, c.Name())
	if len(c.BashCompletionFunction) > 0 {
		buf.WriteString(c.BashCompletionFunction + "\n")
	}
	gen(buf, c)
	writePostscript(buf, c.Name())

	_, err := buf.WriteTo(w)
	return err
}

func nonCompletableFlag(flag *pflag.Flag) bool {
	return flag.Hidden || len(flag.Deprecated) > 0
}

// GenBashCompletionFile generates bash completion ***REMOVED***le.
func (c *Command) GenBashCompletionFile(***REMOVED***lename string) error {
	outFile, err := os.Create(***REMOVED***lename)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return c.GenBashCompletion(outFile)
}

// MarkFlagRequired adds the BashCompOneRequiredFlag annotation to the named flag if it exists,
// and causes your command to report an error if invoked without the flag.
func (c *Command) MarkFlagRequired(name string) error {
	return MarkFlagRequired(c.Flags(), name)
}

// MarkPersistentFlagRequired adds the BashCompOneRequiredFlag annotation to the named persistent flag if it exists,
// and causes your command to report an error if invoked without the flag.
func (c *Command) MarkPersistentFlagRequired(name string) error {
	return MarkFlagRequired(c.PersistentFlags(), name)
}

// MarkFlagRequired adds the BashCompOneRequiredFlag annotation to the named flag if it exists,
// and causes your command to report an error if invoked without the flag.
func MarkFlagRequired(flags *pflag.FlagSet, name string) error {
	return flags.SetAnnotation(name, BashCompOneRequiredFlag, []string{"true"})
}

// MarkFlagFilename adds the BashCompFilenameExt annotation to the named flag, if it exists.
// Generated bash autocompletion will select ***REMOVED***lenames for the flag, limiting to named extensions if provided.
func (c *Command) MarkFlagFilename(name string, extensions ...string) error {
	return MarkFlagFilename(c.Flags(), name, extensions...)
}

// MarkFlagCustom adds the BashCompCustom annotation to the named flag, if it exists.
// Generated bash autocompletion will call the bash function f for the flag.
func (c *Command) MarkFlagCustom(name string, f string) error {
	return MarkFlagCustom(c.Flags(), name, f)
}

// MarkPersistentFlagFilename adds the BashCompFilenameExt annotation to the named persistent flag, if it exists.
// Generated bash autocompletion will select ***REMOVED***lenames for the flag, limiting to named extensions if provided.
func (c *Command) MarkPersistentFlagFilename(name string, extensions ...string) error {
	return MarkFlagFilename(c.PersistentFlags(), name, extensions...)
}

// MarkFlagFilename adds the BashCompFilenameExt annotation to the named flag in the flag set, if it exists.
// Generated bash autocompletion will select ***REMOVED***lenames for the flag, limiting to named extensions if provided.
func MarkFlagFilename(flags *pflag.FlagSet, name string, extensions ...string) error {
	return flags.SetAnnotation(name, BashCompFilenameExt, extensions)
}

// MarkFlagCustom adds the BashCompCustom annotation to the named flag in the flag set, if it exists.
// Generated bash autocompletion will call the bash function f for the flag.
func MarkFlagCustom(flags *pflag.FlagSet, name string, f string) error {
	return flags.SetAnnotation(name, BashCompCustom, []string{f})
}
