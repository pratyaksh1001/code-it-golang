package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/genai"
)

type Problem struct {
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Tags         []string `json:"tags"`
	Languages    []string `json:"languages"`
	DriverCodes  []string `json:"driver_codes"`
	SampleInput  []string `json:"sample_input"`
	SampleOutput []string `json:"sample_output"`
}

type Result struct {
	GoMain          string `json:"go_main"`
	PYMain          string `json:"py_main"`
	JSMain          string `json:"js_main"`
	GoSolution      string `json:"go_solution"`
	PYSolution      string `json:"py_solution"`
	JSSolution      string `json:"js_solution"`
	GoImports       string `json:"go_imports"`
	PYImports       string `json:"py_imports"`
	JsImports       string `json:"js_imports"`
	GoFuncSignature string `json:"go_func_signature"`
	PYFuncSignature string `json:"py_func_signature"`
	JsFuncSignature string `json:"js_func_signature"`
}

var ai *genai.Client

func connect_gemini() {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		fmt.Println("failed to create client: %w", err)
	}
	ai = client
}

func get_gemini_question(topics []string, difficulty string, num_cases int) Problem {
	ctx := context.Background()

	geminiResSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"title": map[string]any{
				"type": "string",
			},
			"description": map[string]any{
				"type": "string",
			},
			"tags": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
			},
			"languages": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
			},
			"driver_codes": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
			},
			"solution_codes": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
			},
			"sample_input": map[string]any{
				"type": "array",
			},
			"sample_output": map[string]any{
				"type": "array",
			},
		},
		"required": []string{
			"title",
			"description",
			"tags",
			"languages",
			"driver_codes",
			"sample_input",
			"sample_output",
			"solution_codes",
		},
	}

	config := &genai.GenerateContentConfig{
		ResponseMIMEType:   "application/json",
		ResponseJsonSchema: geminiResSchema,
	}

	resp, err := ai.Models.GenerateContent(
		ctx,
		"gemini-3.5-flash",
		genai.Text("topics are - "+fmt.Sprint(topics)+"difficulty level - "+difficulty+"number of test cases - "+fmt.Sprint(num_cases)+"language are go,python"+"you have to generate the content that all is compatible with everything the driver code should work properly with sample test cases"),
		config,
	)
	if err != nil {
		fmt.Println("generate content failed: %w", err)
	}

	var problem Problem

	if err := json.Unmarshal([]byte(resp.Text()), &problem); err != nil {
		fmt.Println("failed to unmarshal response: %w", err)
	}

	fmt.Println("\nParsed Struct:")
	fmt.Printf("%+v\n", problem)

	fmt.Println("\nPretty JSON:")
	pretty, _ := json.MarshalIndent(problem, "", "  ")
	fmt.Println(string(pretty))

	return problem

}

func generate_driver_code_from_IO(qid int) {
	var (
		input  string
		output string
	)

	var result Result
	var description string

	db.QueryRow(context.Background(), "select input,output from testcases where qid=$1 limit 1;", qid).Scan(&input, &output)
	db.QueryRow(context.Background(), "select description from question where qid=$1;", qid).Scan(&description)
	driver_generator_schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"go_imports": map[string]any{
				"type": "string",
			},
			"py_imports": map[string]any{
				"type": "string",
			},
			"js_imports": map[string]any{
				"type": "string",
			},
			"go_main": map[string]any{
				"type": "string",
			},
			"py_main": map[string]any{
				"type": "string",
			},
			"js_main": map[string]any{
				"type": "string",
			},
			"go_solution": map[string]any{
				"type": "string",
			},
			"py_solution": map[string]any{
				"type": "string",
			},
			"js_solution": map[string]any{
				"type": "string",
			},
			"go_func_signature": map[string]any{
				"type": "string",
			},
			"py_func_signature": map[string]any{
				"type": "string",
			},
			"js_func_signature": map[string]any{
				"type": "string",
			},
		},
	}
	config := &genai.GenerateContentConfig{
		ResponseMIMEType:   "application/json",
		ResponseJsonSchema: driver_generator_schema,
	}
	var prompt string = ""
	prompt += "do not use escape sequence and give the package and imports in import section only and the solution in solution part and main function that handles all IO and result printing in main part, and function definition with empty code and just function signature that has arguments passed and expected result datatype in the signature section separately and read the question and sample Input and output properly"
	prompt += "description - " + description
	prompt += "input will be like this - " + input
	prompt += "output will be like this - " + output

	resp, err := ai.Models.GenerateContent(context.Background(), "gemini-3.5-flash", genai.Text(prompt), config)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("driver code generation failed")
	}

	err = json.Unmarshal([]byte(resp.Text()), &result)
	if err != nil {
		fmt.Println("json parsing of driver result failed")
	}

	fmt.Println(result.GoMain)
	go db.Exec(context.Background(), "insert into driver_go(qid,main,solution,imports,signature) values($1,$2,$3,$4,$5);", qid, string(result.GoMain), string(result.GoSolution), string(result.GoImports), string(result.GoFuncSignature))
	go db.Exec(context.Background(), "insert into driver_py(qid,main,solution,imports,signature) values($1,$2,$3,$4,$5);", qid, string(result.PYMain), string(result.PYSolution), string(result.PYImports), string(result.PYFuncSignature))
	go db.Exec(context.Background(), "insert into driver_js(qid,main,solution,imports,signature) values($1,$2,$3,$4,$5);", qid, string(result.JSMain), string(result.JSSolution), string(result.JsImports), string(result.JsFuncSignature))
}

func AI_question_gen(c *gin.Context) {
	var data struct {
		Topics     []string `json:"topics"`
		Difficulty string   `json:"difficulty"`
		Num_cases  int      `json:"num_cases"`
	} = struct {
		Topics     []string `json:"topics"`
		Difficulty string   `json:"difficulty"`
		Num_cases  int      `json:"num_cases"`
	}{}
	c.ShouldBindJSON(&data)
	res := get_gemini_question(data.Topics, data.Difficulty, data.Num_cases)
	fmt.Println(res)
	c.JSON(http.StatusOK, gin.H{
		"result": res,
	})
}
