/*
MIT License

Copyright (c) 2023-2024 Rick

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package pkg

import "fmt"

type RepoInformation struct {
	Provider string
	Server   string
	Owner    string
	Repo     string
	PrNumber int

	Username, Token string

	Status      string
	Target      string
	Label       string
	Description string
}

// String returns the human-readable string
func (r RepoInformation) String() string {
	return r.GetRepoPath()
}

// GetRepoPath returns the repository path
func (r RepoInformation) GetRepoPath() string {
	return fmt.Sprintf("%s/%s", r.Owner, r.Repo)
}
