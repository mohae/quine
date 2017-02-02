# quine
Quine isn't actually a quine.

In William Gibson's "Burning Chrome", Bobby Quine is a software expert and part of the duo that burns Chrome.

Quine generates a basic `main.go` for Go cli application. The generated main function will call `appMain()`, which is expected to have the signature:

    func appMain() int {
		// your code
	}

The application's `appMain` should be in a separate file. All custom code, non-quine generated, should be in non-main.go files. This allows for regeneration of `main.go` without affecting other Go code in the `package main`. An example of when `main.go` might be regenerated is after defining new flags.

If quine is to generate flags for the new application, a file with the flag information must be provided. By default, quine looks for `flags.json` in the working directory. To have quine use a different file, a filename must be provided using the `flag` flag.

Quine will include the license file as specified by the `-license` flag. Either a copy of the license, or the license notice text, if the license has such text, will be added to `main.go`. If the notice text includes fields that should be replaced with the application and author's information, the replacement will be done, if quine has the information. GPL licenses also have license information for CLIs which will be displayed by the application when it starts. When an application uses a GPL license, the flags referenced by the CLI license information will be added to the application's flags, along with the functions to support the flags.


## Usage
Generate an application named foo:

    $ quine foo


Generate an application named foo

## Flags
