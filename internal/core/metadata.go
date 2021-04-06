package core

import (
	"bufio"
	"bytes"
	"net/textproto"
	"strings"

	"google.golang.org/grpc/metadata"
)

type MetadataFlags metadata.MD

func (m *MetadataFlags) String() string {
	return ""
}

func (m *MetadataFlags) Set(s string) error {
	h, err := textproto.NewReader(bufio.NewReader(bytes.NewBufferString(s + "\n\n"))).ReadMIMEHeader()
	if err != nil {
		return err
	}
	for k, v := range h {
		delete(h, k)
		h[strings.ToLower(k)] = v
	}
	*m = MetadataFlags(metadata.Join(metadata.MD(*m), metadata.MD(h)))
	return nil
}

func (m *MetadataFlags) Type() string {
	return "test"
}

func (m *MetadataFlags) MD() metadata.MD {
	return metadata.MD(*m)
}
