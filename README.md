# cge-parser
![CGE Version](https://img.shields.io/badge/CGE-v0.5-green)

A parser for the [CodeGame Events Language](https://code-game.org/docs/specification/cge).

## Usage

*cge-parser* is not meant to be used as a standalone program.

*cge-parser* receives the CGE file over STDIN and sends its output as [protobuf](https://protobuf.dev/) messages over STDOUT.

### Flags

- `--comments`: include doc comments in output
- `--only-meta`: stop parsing after sending the metadata message
- `--tokens`: return all parsed tokens
- `--no-objects`: do not return objects
- `--no-warn`: disable warnings

### Output messages

See [schema.proto](./protobuf/schema.proto).

## License

Copyright (C) 2023 Julian Hofmann

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
