/*
given a compatible http server, `httpfile` exposes HTTP endpoint as a file
manipulation tool that's satisfy io.ReadWriteCloser and .Stat()

Required HTTP endpoint is
- read the file and return the stream of file content
- write to file given a stream of file content
- return file info as json-encoded data
*/
package httpfile
