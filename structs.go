package main

type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Login_data struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Question struct {
	Problem string   `json:"problem"`
	Tags    []string `json:"tags"`
	Email   string   `json:"email"`
	Title   string   `json:"title"`
}

type TestCase struct {
	Qid    int    `json:"qid"`
	Input  string `json:"input"`
	Output string `json:"output"`
}
