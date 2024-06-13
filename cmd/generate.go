package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func isFileExist(filePath string) bool {
	_, error := os.Stat(filePath)
	//return !os.IsNotExist(err)
	return !errors.Is(error, os.ErrNotExist)
}

func firstLetterToLower(s string) string {

	if len(s) == 0 {
		return s
	}

	r := []rune(s)
	r[0] = unicode.ToLower(r[0])

	return string(r)
}

func Execute() {

	var template []byte
	var input []byte
	var err error

	var rootCmd = &cobra.Command{
		Use:   "grog",
		Short: "Grog is a very fast .net code generator",
		Long:  `A Fast and Flexible .net code Generator built with love by aleluis in Go.`,
		Run: func(cmd *cobra.Command, args []string) {

			fmt.Print("🙈 Name of project*: ")
			reader := bufio.NewReader(os.Stdin)
			project, _ := reader.ReadString('\n')
			project = strings.TrimRight(project, "\r\n")

			fmt.Print("🙉 Name of the Controller*: ")
			reader = bufio.NewReader(os.Stdin)
			name, _ := reader.ReadString('\n')
			name = strings.TrimRight(name, "\r\n")

			fmt.Print("🙊 Name of use case (empty to skip): ")
			reader = bufio.NewReader(os.Stdin)
			useCase, _ := reader.ReadString('\n')
			useCase = strings.TrimRight(useCase, "\r\n")

			// Program add Services
			if isFileExist("./Program.cs") {
				input, err = os.ReadFile("Program.cs")
				check(err)

				template = []byte(ProgramTpl(name, useCase))

				output := bytes.Replace(input, []byte("// Add services to the container."), template, -1)

				if err = os.WriteFile("./Program.cs", output, 0666); err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}

			// Controller

			path := "./Infrastructure/Controllers"
			err := os.MkdirAll(path, os.ModePerm)
			check(err)

			if !isFileExist("./Infrastructure/Controllers/" + name + "Controller.cs") {
				d1 := []byte(ControllerTpl(name, project))
				err = os.WriteFile("./Infrastructure/Controllers/"+name+"Controller.cs", d1, 0644)
				check(err)
			}

			// Use Case - if not empty

			if useCase != "" {

				// Interface Use Case

				path = "./Domain/Ports/Input"
				err = os.MkdirAll(path, os.ModePerm)
				check(err)

				if !isFileExist("./Domain/Ports/Input/I" + useCase + "UseCase.cs") {
					template = []byte(InterfaceUseCaseTpl(useCase, project))
					err = os.WriteFile("./Domain/Ports/Input/I"+useCase+"UseCase.cs", template, 0644)
					check(err)
				}

				// Use Case

				path = "./Application/UseCases"
				err = os.MkdirAll(path, os.ModePerm)
				check(err)

				if !isFileExist("./Application/UseCases/" + useCase + "UseCase.cs") {
					template = []byte(UseCaseTpl(useCase, project))
					err = os.WriteFile("./Application/UseCases/"+useCase+"UseCase.cs", template, 0644)
					check(err)
				}

			}

			// Interface Service

			path = "./Domain/Ports/Input"
			err = os.MkdirAll(path, os.ModePerm)
			check(err)

			if !isFileExist("./Domain/Ports/Input/I" + name + "Service.cs") {
				template = []byte(InterfaceServiceTpl(name, project, useCase))
				err = os.WriteFile("./Domain/Ports/Input/I"+name+"Service.cs", template, 0644)
				check(err)
			}

			// Service

			path = "./Application/Services"
			err = os.MkdirAll(path, os.ModePerm)
			check(err)

			if !isFileExist("./Application/Services/" + name + "Service.cs") {
				template = []byte(ServiceTpl(name, project, useCase))
				err = os.WriteFile("./Application/Services/"+name+"Service.cs", template, 0644)
				check(err)
			}

			fmt.Println("🎉 Done!")

		},
	}

	//rootCmd.Flags().StringVarP(&name, "name", "n", "", "")
	//rootCmd.Flags().StringVarP(&proyecto, "proyecto", "p", "", "")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func ControllerTpl(name string, proyecto string) string {
	return fmt.Sprintf(`
using Microsoft.AspNetCore.Mvc;
using %[2]v.Domain.Ports.Input;

namespace %[2]v.Infrastructure.Controllers;

[ApiController]
[Route("api/[controller]")]
public class %[1]vController : ControllerBase
{
    private readonly ILogger<%[1]vController> _logger;
	private readonly I%[1]vService _service;

    public %[1]vController(ILogger<%[1]vController> logger, I%[1]vService service)
    {
        _logger = logger;
		_service = service;
    }

    /// <summary>
    /// Metodo Ping de prueba
    /// </summary>
    /// <returns>pong</returns>
    [HttpGet("")]
    public ActionResult<string> Ping() 
    {
        try {
            _logger.LogWarning("Se ha ejecutado el método Ping");
            return Ok($"pong {DateTime.Now.ToString("yyyy-MM-dd hh:mm:ss")}");
        }
        catch (Exception e)
        {
            _logger.LogError(e, "Se produjo un error en el método Ping");
            return Problem(e.Message);
        }
    }

}
	`, name, proyecto)
}

func ServiceTpl(name string, proyecto string, useCase string) string {

	if useCase == "" {
		return fmt.Sprintf(`
using NLog;
using %[2]v.Domain.Ports.Input;

namespace %[2]v.Application.Services;
public class %[1]vService : I%[1]vService
{
	private Logger _logger = LogManager.GetCurrentClassLogger();
	public %[1]vService()
	{
	}
}`, name, proyecto, useCase)
	} else {
		return fmt.Sprintf(`
using NLog;
using %[2]v.Domain.Ports.Input;

namespace %[2]v.Application.Services;
public class %[1]vService : I%[1]vService
{
	private Logger _logger = LogManager.GetCurrentClassLogger();
	private I%[3]vUseCase _%[4]vUseCase;
	public %[1]vService(I%[3]vUseCase %[4]vUseCase)
	{
		_%[4]vUseCase = %[4]vUseCase;
	}
}`, name, proyecto, useCase, firstLetterToLower(useCase))
	}
}

func InterfaceServiceTpl(name string, proyecto string, useCase string) string {

	if useCase != "" {
		return fmt.Sprintf(`
namespace %[2]v.Domain.Ports.Input;
public interface I%[1]vService : I%[3]vUseCase
{
}
		`, name, proyecto, useCase)
	} else {

		return fmt.Sprintf(`
namespace %[2]v.Domain.Ports.Input;
public interface I%[1]vService
{
}
		`, name, proyecto)
	}
}

func UseCaseTpl(useCase string, proyecto string) string {
	return fmt.Sprintf(`
using NLog;
using %[2]v.Domain.Ports.Input;

namespace %[2]v.Application.UseCases;
public class %[1]vUseCase : I%[1]vUseCase
{
	private Logger _logger = LogManager.GetCurrentClassLogger();
	public %[1]vUseCase()
	{
	}
}
	`, useCase, proyecto)
}

func InterfaceUseCaseTpl(useCase string, proyecto string) string {
	return fmt.Sprintf(`
namespace %[2]v.Domain.Ports.Input;
public interface I%[1]vUseCase
{
}
	`, useCase, proyecto)
}

func ProgramTpl(name string, useCase string) string {

	if useCase == "" {
		return fmt.Sprintf(`// Add services to the container.
builder.Services.AddScoped<I%[1]vService, %[1]vService>();`, name, useCase)
	} else {
		return fmt.Sprintf(`// Add services to the container.
builder.Services.AddScoped<I%[2]vUseCase, %[2]vUseCase>();
builder.Services.AddScoped<I%[1]vService, %[1]vService>();`, name, useCase)
	}

}
