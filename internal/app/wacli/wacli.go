// 版权 @2019 凹语言 作者。保留所有权利。

//go:build !wasm
// +build !wasm

// 凹语言，The Wa Programming Language.
package wacli

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/tetratelabs/wazero/sys"

	"wa-lang.org/wa/api"
	"wa-lang.org/wa/internal/3rdparty/cli"
	"wa-lang.org/wa/internal/app"
	"wa-lang.org/wa/internal/app/apputil"
	"wa-lang.org/wa/internal/app/yacc"
	"wa-lang.org/wa/internal/config"
)

func Main() {
	cliApp := cli.NewApp()
	cliApp.Name = "Wa"
	cliApp.Usage = "Wa is a tool for managing Wa source code."
	cliApp.Version = func() string {
		if info, ok := debug.ReadBuildInfo(); ok {
			if info.Main.Version != "" {
				return info.Main.Version
			}
		}
		return "devel:" + time.Now().Format("2006-01-02+15:04:05")
	}()

	cliApp.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "target",
			Usage: "set target os (arduino|chrome|wasi)",
			Value: config.WaOS_Default,
		},
		&cli.BoolFlag{
			Name:    "debug",
			Aliases: []string{"d"},
			Usage:   "set debug mode",
		},
		&cli.StringFlag{
			Name:    "trace",
			Aliases: []string{"t"},
			Usage:   "set trace mode (*|app|compiler|loader)",
		},
	}

	cliApp.Before = func(c *cli.Context) error {
		switch c.String("target") {
		case "wasi", "arduino", "chrome":
			// OK
		default:
			fmt.Printf("unknown target: %s\n", c.String("target"))
			os.Exit(1)
		}
		if c.Bool("debug") {
			config.SetDebugMode()
		}
		if parten := c.String("trace"); parten != "" {
			config.SetEnableTrace(parten)
		}
		return nil
	}

	// 没有参数时对应 run 命令
	cliApp.Action = func(c *cli.Context) error {
		cliRun(c)
		return nil
	}

	cliApp.Commands = []*cli.Command{
		{
			// go run main.go debug
			Hidden: true,
			Name:   "debug",
			Usage:  "only for dev/debug",
			Action: func(c *cli.Context) error {
				wat, err := api.BuildFile(
					config.DefaultConfig(),
					"hello.wa", "func main() { println(123) }",
				)
				if err != nil {
					if len(wat) != 0 {
						fmt.Println(string(wat))
					}
					fmt.Println(err)
					os.Exit(1)
				}
				fmt.Println(string(wat))
				return nil
			},
		},
		{
			Name:      "init",
			Usage:     "init a sketch wa module",
			ArgsUsage: "app-name",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "name",
					Aliases: []string{"n"},
					Usage:   "set app name",
					Value:   "_examples/hello",
				},
				&cli.StringFlag{
					Name:    "pkgpath",
					Aliases: []string{"p"},
					Usage:   "set pkgpath file",
					Value:   "myapp",
				},
				&cli.BoolFlag{
					Name:    "update",
					Aliases: []string{"u"},
					Usage:   "update example",
				},
			},

			Action: func(c *cli.Context) error {
				waApp := app.NewApp(build_Options(c))
				err := waApp.InitApp(c.String("name"), c.String("pkgpath"), c.Bool("update"))
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				return nil
			},
		},
		{
			Name:  "run",
			Usage: "compile and run Wa program",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "target",
					Usage: "set target os (wasi|arduino|chrome)",
					Value: config.WaOS_Default,
				},
				&cli.StringFlag{
					Name:  "tags",
					Usage: "set build tags",
				},
			},
			Action: func(c *cli.Context) error {
				cliRun(c)
				return nil
			},
		},
		{
			Name:  "build",
			Usage: "compile Wa source code",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "output",
					Aliases: []string{"o"},
					Usage:   "set output file",
					Value:   "a.out.wasm",
				},
				&cli.StringFlag{
					Name:  "target",
					Usage: "set target os (wasi|arduino|chrome)",
					Value: config.WaOS_Default,
				},
				&cli.StringFlag{
					Name:  "tags",
					Usage: "set build tags",
				},
				&cli.IntFlag{
					Name:  "ld-stack-size",
					Usage: "set stack size",
				},
				&cli.IntFlag{
					Name:  "ld-max-memory",
					Usage: "set max memory size",
				},
			},
			Action: func(c *cli.Context) error {
				outfile := c.String("output")

				if c.NArg() == 0 {
					fmt.Fprintf(os.Stderr, "no input file")
					os.Exit(1)
				}

				ctx := app.NewApp(build_Options(c))
				output, err := ctx.WASM(c.Args().First())
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				if outfile != "" && outfile != "-" {
					watFilename := outfile
					if !strings.HasSuffix(watFilename, ".wat") {
						watFilename += ".wat"
					}
					err := os.WriteFile(watFilename, []byte(output), 0666)
					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
					if strings.HasSuffix(outfile, ".wasm") {
						if stdoutStderr, err := apputil.RunWat2Wasm(watFilename, "-o", outfile); err != nil {
							fmt.Println(string(stdoutStderr))
							os.Exit(1)
						}
						os.Remove(watFilename)
					}
				} else {
					fmt.Println(string(output))
				}

				return nil
			},
		},
		{
			Hidden: true,
			Name:   "native",
			Usage:  "compile wa source code to native executable",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "output",
					Aliases: []string{"o"},
					Usage:   "set output file",
					Value:   "",
				},
				&cli.StringFlag{
					Name:  "target",
					Usage: "set native target",
					Value: "",
				},
				&cli.StringFlag{
					Name:  "tags",
					Usage: "set build tags",
				},
				&cli.BoolFlag{
					Name:  "debug",
					Usage: "dump orginal intermediate representation",
				},
				&cli.StringFlag{
					Name:  "clang",
					Usage: "set llvm/clang path",
				},
				&cli.StringFlag{
					Name:  "llc",
					Usage: "set llvm/llc path",
				},
			},
			Action: func(c *cli.Context) error {
				outfile := c.String("output")
				target := c.String("target")
				debug := c.Bool("debug")
				infile := ""

				if c.NArg() == 0 {
					fmt.Fprintf(os.Stderr, "no input file")
					os.Exit(1)
				}
				infile = c.Args().First()

				ctx := app.NewApp(build_Options(c, config.WaBackend_llvm))
				if err := ctx.LLVM(infile, outfile, target, debug); err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				return nil
			},
		},

		{
			Hidden: true,
			Name:   "lex",
			Usage:  "lex Wa source code and print token list",
			Action: func(c *cli.Context) error {
				if c.NArg() == 0 {
					fmt.Fprintf(os.Stderr, "no input file")
					os.Exit(1)
				}

				waApp := app.NewApp(build_Options(c))
				err := waApp.Lex(c.Args().First())
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				return nil
			},
		},
		{
			Hidden: true,
			Name:   "ast",
			Usage:  "parse Wa source code and print ast",
			Action: func(c *cli.Context) error {
				if c.NArg() == 0 {
					fmt.Fprintf(os.Stderr, "no input file")
					os.Exit(1)
				}

				waApp := app.NewApp(build_Options(c))
				err := waApp.AST(c.Args().First())
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				return nil
			},
		},
		{
			Hidden: true,
			Name:   "ssa",
			Usage:  "print Wa ssa code",
			Action: func(c *cli.Context) error {
				if c.NArg() == 0 {
					fmt.Fprintf(os.Stderr, "no input file")
					os.Exit(1)
				}

				ctx := app.NewApp(build_Options(c))
				err := ctx.SSA(c.Args().First())
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				return nil
			},
		},
		{
			Hidden: true,
			Name:   "cir",
			Usage:  "print cir code",
			Action: func(c *cli.Context) error {
				if c.NArg() == 0 {
					fmt.Fprintf(os.Stderr, "no input file")
					os.Exit(1)
				}

				ctx := app.NewApp(build_Options(c))
				err := ctx.CIR(c.Args().First())
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				return nil
			},
		},
		{
			Hidden: true,
			Name:   "test",
			Usage:  "test packages",
			Action: func(c *cli.Context) error {
				if c.NArg() < 1 {
					cli.ShowAppHelpAndExit(c, 0)
				}
				appArgs := c.Args().Slice()[1:]
				waApp := app.NewApp(build_Options(c))
				err := waApp.RunTest(c.Args().First(), appArgs...)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				return nil
			},
		},
		{
			Name:  "fmt",
			Usage: "format Wa package sources",
			Action: func(c *cli.Context) error {
				waApp := app.NewApp(build_Options(c))
				err := waApp.Fmt(c.Args().First())
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				return nil
			},
		},
		{
			Hidden: true,
			Name:   "doc",
			Usage:  "show documentation for package or symbol",
			Action: func(c *cli.Context) error {
				fmt.Println("TODO")
				return nil
			},
		},
		{
			Hidden: true,
			Name:   "install-wat2wasm",
			Usage:  "install-wat2wasm tool",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "dir",
					Usage: "set output dir",
					Value: "",
				},
			},
			Action: func(c *cli.Context) error {
				outdir := c.String("dir")
				if err := apputil.InstallWat2wasm(outdir); err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				return nil
			},
		},

		{
			Name:      "yacc",
			Usage:     "generates parsers for LALR(1) grammars",
			ArgsUsage: "<input>",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "l",
					Usage: "disable line directives",
				},
				&cli.StringFlag{
					Name:  "o",
					Usage: "set parser output file",
					Value: "y.wa",
				},
				&cli.StringFlag{
					Name:  "p",
					Usage: "name prefix to use in generated code",
					Value: "yy",
				},
				&cli.StringFlag{
					Name:  "v",
					Usage: "create parsing tables",
					Value: "y.output",
				},
				&cli.StringFlag{
					Name:  "c",
					Usage: "set copyright file",
					Value: "",
				},
			},
			Action: func(c *cli.Context) error {
				if c.NArg() != 1 {
					cli.ShowSubcommandHelpAndExit(c, 1)
				}
				yacc.InitFlags(yacc.Flags{
					Oflag:     c.String("o"),
					Vflag:     c.String("v"),
					Lflag:     c.Bool("l"),
					Prefix:    c.String("p"),
					Copyright: loadCopyright(c.String("c")),
				})
				yacc.Main(c.Args().First())
				return nil
			},
		},

		{
			Name:  "logo",
			Usage: "print wa-lang text format logo",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "more",
					Usage: "print more logos",
				},
			},
			Action: func(c *cli.Context) error {
				app.PrintLogo(c.Bool("more"))
				return nil
			},
		},
	}

	cliApp.Run(os.Args)
}

func build_Options(c *cli.Context, waBackend ...string) *app.Option {
	opt := &app.Option{
		Debug:        c.Bool("debug"),
		WaBackend:    config.WaBackend_Default,
		BuilgTags:    strings.Fields(c.String("tags")),
		Clang:        c.String("clang"),
		Llc:          c.String("llc"),
		LD_StackSize: c.Int("ld-stack-size"),
		LD_MaxMemory: c.Int("ld-max-memory"),
	}

	opt.TargetArch = "wasm"
	if len(waBackend) > 0 {
		opt.WaBackend = waBackend[0]
	}
	switch c.String("target") {
	case "", "wa", "walang":
		opt.TargetOS = config.WaOS_Default
	case config.WaOS_wasi:
		opt.TargetOS = config.WaOS_wasi
	case config.WaOS_arduino:
		opt.TargetOS = config.WaOS_arduino
	case config.WaOS_chrome:
		opt.TargetOS = config.WaOS_chrome
	default:
		fmt.Printf("unknown target: %s\n", c.String("target"))
		os.Exit(1)
	}
	return opt
}

func cliRun(c *cli.Context) {
	if c.NArg() < 1 {
		cli.ShowAppHelpAndExit(c, 0)
	}

	var app = app.NewApp(build_Options(c))
	var infile = c.Args().First()
	var outfile string
	var output []byte
	var err error

	if !c.Bool("debug") {
		defer os.Remove(outfile)
	}

	switch {
	case strings.HasSuffix(infile, ".wat"):
		outfile = infile
		output, err = os.ReadFile(infile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case strings.HasSuffix(infile, ".wasm"):
		outfile = infile
		output, err = os.ReadFile(infile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	default:
		outfile = "a.out.wat"
		output, err = app.WASM(infile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if err = os.WriteFile(outfile, []byte(output), 0666); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	appArgs := c.Args().Slice()[1:]
	stdoutStderr, err := apputil.RunWasm(app.GetConfig(), outfile, appArgs...)
	if err != nil {
		if len(stdoutStderr) > 0 {
			fmt.Println(string(stdoutStderr))
		}
		if exitErr, ok := err.(*sys.ExitError); ok {
			os.Exit(int(exitErr.ExitCode()))
		}
		fmt.Println(err)
	}
	if len(stdoutStderr) > 0 {
		fmt.Println(string(stdoutStderr))
	}
}

func loadCopyright(filename string) string {
	data, _ := os.ReadFile(filename)
	return string(data)
}
