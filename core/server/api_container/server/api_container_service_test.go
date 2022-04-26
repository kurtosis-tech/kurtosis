/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package server

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis-core/api/golang/kurtosis_core_rpc_api_bindings"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

func TestTarFiles(t *testing.T) {
	// Create and add some files to the archive.
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	var files = []struct {
		Name, Body string
	}{
		{"readme.txt", "This archive contains some text files."},
		{"gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"},
		{"todo.txt", "Get animal handling license."},
	}
	for _, file := range files {
		hdr := &tar.Header{
			Name: file.Name,
			Mode: 0600,
			Size: int64(len(file.Body)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			log.Fatal(err)
		}
		if _, err := tw.Write([]byte(file.Body)); err != nil {
			log.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		log.Fatal(err)
	}

	// Open and iterate through the files in the archive.
	tr := tar.NewReader(&buf)

	destination := "/Users/lporoli/Kurtosis-Test/enclave-data"
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		hdrInfo := hdr.FileInfo()

		dstpath := path.Join(destination, hdr.Name)
		// Overriding permissions to allow writing content
		OWNER_PERM_RW := os.FileMode(0600)
		mode := hdrInfo.Mode() | OWNER_PERM_RW

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(dstpath, mode); err != nil {
				if !os.IsExist(err) {
					log.Fatal(err)
				}
				err = os.Chmod(dstpath, mode)
				if err != nil {
					log.Fatal(err)
				}
			}
		case tar.TypeReg, tar.TypeRegA:
			file, err := os.OpenFile(dstpath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
			if err != nil {
				log.Fatal(err)
			}
			if _, err := io.Copy(file, tr); err != nil {
				file.Close()
				log.Fatal(err)
			}
			file.Close()
		case tar.TypeSymlink:
			if err := os.Symlink(hdr.Linkname, dstpath); err != nil {
				log.Fatal(err)
			}
		case tar.TypeLink:
			target := path.Join(destination, strings.TrimPrefix(hdr.Linkname, "bla"))
			if err := os.Link(target, dstpath); err != nil {
				log.Fatal(err)
			}
		default:
			// For now we're skipping anything else. Special device files and
			// symlinks are not needed or anyway probably incorrect.
		}

		// maintaining access and modification time in best effort fashion
		os.Chtimes(dstpath, hdr.AccessTime, hdr.ModTime)
	}

	var tgbuf bytes.Buffer
	err := compress(destination, &tgbuf)
	if err != nil {
		log.Fatal(err)
	}

	destination2 := "/Users/lporoli/Kurtosis-Test/otro"
	// write the .tar.gzip
	fileToWrite, err := os.OpenFile(destination2 +"/pruebilita.tgz", os.O_CREATE|os.O_RDWR, os.FileMode(0777))
	if err != nil {
		panic(err)
	}
	if _, err := io.Copy(fileToWrite, &buf); err != nil {
		panic(err)
	}

	/*
		var wholeContentByte []byte
		for {
			// hdr gives you the header of the tar file
			hdr, err := tr.Next()
			if err == io.EOF {
				// end of tar archive
				break
			}
			if err != nil {
				log.Fatalln(err)
			}

			buf := new(bytes.Buffer)
			buf.ReadFrom(tr)

			// You can use this wholeContent to create new file
			wholeContent := buf.String()
			wholeContentByte = append(wholeContentByte,  buf.Bytes()...)

			fmt.Println("Whole of the string of ", hdr.Name ," is ",wholeContent)
		}

		gbuf := new(bytes.Buffer)

		w := gzip.NewWriter(gbuf)
		w.Name = "pruebanueva"
		if _, err := w.Write(wholeContentByte); err != nil {
			t.Fatalf("Write: %v", err)
		}
		if err := w.Close(); err != nil {
			t.Fatalf("Writer.Close: %v", err)
		}

		log.Println("Algo al final")

		enclaveDataDirectory := enclave_data_directory.NewEnclaveDataDirectory("/Users/lporoli/Kurtosis-Test/enclave-data")

		artifactCache, err := enclaveDataDirectory.GetFilesArtifactCache()
		if err != nil {
			log.Fatal(err)
		}

		bytesReader := bytes.NewReader(gbuf.Bytes())

		uuid, err := artifactCache.StoreFile(bytesReader, "pruebilla")
		if err != nil {
			log.Fatal(err)
		}
		log.Println("The new uuid is: "+ uuid)

	*/

	/*uuid, err := artifactCache.StoreFile(tr, "probando.tar")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("New uuid: "+uuid)*/

	/*for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Contents of %s:\n", hdr.Name)
		if _, err := io.Copy(os.Stdout, tr); err != nil {
			log.Fatal(err)
		}
		fmt.Println()
	}*/

	// Output:
	// Contents of readme.txt:
	// This archive contains some text files.
	// Contents of gopher.txt:
	// Gopher names:
	// George
	// Geoffrey
	// Gonzo
	// Contents of todo.txt:
	// Get animal handling license.
}

func compress(src string, buf io.Writer) error {
	// tar > gzip > buf
	zr := gzip.NewWriter(buf)
	zr.Close()
	tw := tar.NewWriter(zr)
	tw.Close()

	// walk through every file in the folder
	filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
		// generate tar header
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		// must provide real name
		// (see https://golang.org/src/archive/tar/common.go?#L626)
		header.Name = filepath.ToSlash(file)

		// write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		// if not a dir, write file content
		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, data); err != nil {
				return err
			}
		}
		return nil
	})

	return nil
}

func TestOneToOneApiAndPortSpecProtoMapping(t *testing.T) {
	// Ensure all port spec protos are covered
	require.Equal(t, len(kurtosis_core_rpc_api_bindings.Port_Protocol_name), len(apiContainerPortProtoToPortSpecPortProto))
	for enumInt, enumName := range kurtosis_core_rpc_api_bindings.Port_Protocol_name {
		_, found := apiContainerPortProtoToPortSpecPortProto[kurtosis_core_rpc_api_bindings.Port_Protocol(enumInt)]
		require.True(t, found, "No port spec port proto found for API port proto '%v'", enumName)
	}

	// Ensure no duplicates in the kurtosis backend port protos
	require.Equal(t, len(port_spec.PortProtocolValues()), len(apiContainerPortProtoToPortSpecPortProto))
	seenPortSpecProtos := map[port_spec.PortProtocol]kurtosis_core_rpc_api_bindings.Port_Protocol{}
	for apiPortProto, portSpecProto := range apiContainerPortProtoToPortSpecPortProto {
		preexistingApiPortProto, found := seenPortSpecProtos[portSpecProto]
		require.False(
			t,
			found,
			"port spec proto '%v' is already mapped to API port protocol '%v'",
			portSpecProto,
			preexistingApiPortProto.String(),
		)
		seenPortSpecProtos[portSpecProto] = apiPortProto
	}
}
