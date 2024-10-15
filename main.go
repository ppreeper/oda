/*
Copyright © 2024 Peter Preeper <ppreeper@gmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package main

import (
	_ "embed"

	"github.com/ppreeper/oda/internal"
)

//go:generate sh -c "printf '%s (%s)' $(git tag -l --contains HEAD) $(date +%Y%m%d)-$(git rev-parse --short HEAD)" > commit.txt
//go:embed commit.txt
var Commit string

func main() {
	internal.Execute()
}
