/*
 * Copyright © 2020 Mário Franco
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

package seekerr

import "fmt"

var (
	// commitHash contains the current Git revision.
	commitHash string

	// buildDate contains the date of the current build.
	buildDate string
)

type Version struct {
	// Major
	Major int
	// Minor
	Minor int
	// Increment this for bug releases
	Patch int
	// It will be blank for release versions.
	Suffix string
}

func (v *Version) String() string {
	return fmt.Sprintf("%d.%d.%d%s", v.Major, v.Minor, v.Patch, v.Suffix)
}

type Info struct {
	CommitHash string
	BuildDate string

	// The current version
	Version Version
}

func NewInfo() Info {
	return Info{
		CommitHash: commitHash,
		BuildDate:  buildDate,
		Version: Version{
			Major:  1,
			Minor:  0,
			Patch:  0,
			Suffix: "",
		},
	}
}
