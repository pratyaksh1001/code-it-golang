package main

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/genai"
)

type Problem struct {
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Tags         []string `json:"tags"`
	Languages    []string `json:"languages"`
	DriverCodes  []string `json:"driver_codes"`
	SampleInput  string   `json:"sample_input"`
	SampleOutput string   `json:"sample_output"`
}
type Result struct {
	GoDriver   string `json:"go_driver"`
	PYdriver   string `json:"py_driver"`
	JSDriver   string `json:"js_driver"`
	GoSolution string `json:"go_solution"`
	PYSolution string `json:"py_solution"`
	JSSolution string `json:"js_solution"`
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

func get_gemini_question() Problem {
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
				"type": "string",
			},
			"sample_output": map[string]any{
				"type": "string",
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
		"gemini-2.5-flash",
		genai.Text("Generate an easy graph problem in"),
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
			"go_driver": map[string]any{
				"type": "string",
			},
			"py_driver": map[string]any{
				"type": "string",
			},
			"js_driver": map[string]any{
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
		},
	}
	config := &genai.GenerateContentConfig{
		ResponseMIMEType:   "application/json",
		ResponseJsonSchema: driver_generator_schema,
	}
	var prompt string = ""
	prompt += "DO NOT USE ESCAPE SEQUENCES MAKE SURE THE RESPONSE IS WELL FORMATTED AND INDENTED you have to generate driver code and solution code only in go,js,py and the function that user runs should be named solution and the rest you have to handle do not use escape sequences for anything and the solution code should also strictly take the input and output format in consideration and there should be a function named after the question title and the usetr is supposed to write that function only with all the required arguments passed by driver code the user will not define the function you have to give function defined as well"
	prompt += "description - " + description
	prompt += "input will be like this - " + input
	prompt += "output will be like this - " + output

	resp, err := ai.Models.GenerateContent(context.Background(), "gemini-2.5-flash", genai.Text(prompt), config)
	if err != nil {
		fmt.Println("driver code generation failed")
	}
	//fmt.Println(string(resp.Text()))
	err = json.Unmarshal([]byte(resp.Text()), &result)
	if err != nil {
		fmt.Println("json parsing of driver result failed")
	}
	//pretty, _ := json.MarshalIndent(result, "", "    ")
	fmt.Println(result.GoDriver)
	go db.Exec(context.Background(), "insert into driver_go(qid,code,solution) values($1,$2,$3);", qid, string(result.GoDriver), string(result.GoSolution))
	go db.Exec(context.Background(), "insert into driver_py(qid,code,solution) values($1,$2,$3);", qid, string(result.PYdriver), string(result.PYSolution))
	go db.Exec(context.Background(), "insert into driver_js(qid,code,solution) values($1,$2,$3);", qid, string(result.JSDriver), string(result.JSSolution))
}
