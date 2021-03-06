package BERserk

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"log"
	"math/big"
	"testing"
)

var (
	certData, _ = hex.DecodeString("3082034a30820232a0030201020208670d778714fb84dd300d06092a864886f70d01010b050030323120301e060355040a131746696c6970706f2050776e27696e67204c696d69746564310e300c060355040713054561727468301e170d3135303330383133303435335a170d3230303330363133303935335a30323120301e060355040a131746696c6970706f2050776e27696e67204c696d69746564310e300c06035504071305456172746830820120300d06092a864886f70d01010105000382010d00308201080282010100a6ddac5f80e6a02db689abb363ab23333c2c049f43fa37bb7b442bc7060fbb4d281ac88ba59e655db34e2d6b81509ece5c5b65d092091b9c525d5a8907253c1bc035d0623351e26b447f020f17a71e2ea7bb823d70f1f358c6f817cbfd8f119cbd457eefa8d398790627b0d4b37e9553f3f6bec6078d601a000c23cd8f67e46c556a25d226c693edc5936ab69029847c4d4d5e668dbc4a0b5c49b9fe881998e1982cbd677409263c979077f54d6f17e25b06d6614a462dca1d9d6ae64235ab9164c58eaa86d652f0a0698c665d3f53e7866a0bd203fb17d59c852c0524d15cfa85442259cdef6725591c2e0c9aed38bf5de919c7881fc2718626a023f4dc6767020103a3663064300e0603551d0f0101ff04040302010630120603551d130101ff040830060101ff020102301d0603551d0e0416041495f9365049577c3ac9a9fbccca2461606e631303301f0603551d2304183016801495f9365049577c3ac9a9fbccca2461606e631303300d06092a864886f70d01010b0500038201010081473f2e28744c2623a1ededf994d54aba61b24b643f86766eb2d249f13af42dde7fd54dfe90ee1230f2d075a8965e7f110618f16179df0f1bea3e351c7947aea30c980fdc947bcdf07e6a09c5ee47362897dbc3a8ba4a43078930b4ab558bc4596aa5f6875af3d0931eb5bd842d9513d4b2226491184bc4d15100c1ed1ef751027cd724a0514adcfc3578716cb796a41889d857c2940aca088cc2ac18476170aa829858c7f006ddab678c01de9c6a94624ebe5895c441a78233c15f11777d28e8e4b804ba747a8842c4f92250ba02ea0880ee147cf3bec174ba90565c7de317df1e737d2018977755382798eb364ca14e54cab16b18616894ddb63276a84ae5")
	testCA, _   = x509.ParseCertificate(certData)
)

func TestSign2048(t *testing.T) {
	signer, _, err := New(testCA)
	if err != nil {
		t.Fatal(err)
	}

	var (
		hash = make([]byte, crypto.SHA1.Size())
		sig  []byte
	)
	for {
		_, err = rand.Read(hash)
		if err != nil {
			t.Fatal(err)
		}
		sig, err = signer.Sign(nil, hash, crypto.SHA1)
		if _, ok := err.(ErrRetry); ok {
			t.Log(err)
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		break
	}

	tmpl := RSA2048SHA1DigestInfoTemplate
	cube := new(big.Int).Exp(new(big.Int).SetBytes(sig), THREE, nil).Bytes()
	sigMsg := make([]byte, tmpl.BitLen/8)
	copy(sigMsg[len(sigMsg)-len(cube):], cube)

	if !bytes.Equal(sigMsg[0:len(tmpl.Prefix)], tmpl.Prefix) {
		log.Fatalf("Wrong prefix %x", sigMsg[0:len(tmpl.Prefix)])
	}

	index := tmpl.BitLen/8 - tmpl.HashLen - len(tmpl.Suffix) - tmpl.MiddleOffset - len(tmpl.Middle)
	if !bytes.Equal(sigMsg[index:index+len(tmpl.Middle)], tmpl.Middle) {
		log.Fatalf("Wrong Middle %x", sigMsg[index:index+len(tmpl.Middle)])
	}

	index = tmpl.BitLen/8 - tmpl.HashLen - len(tmpl.Suffix)
	if !bytes.Equal(sigMsg[index:index+len(tmpl.Suffix)], tmpl.Suffix) {
		log.Fatalf("Wrong Suffix %x", sigMsg[index:index+len(tmpl.Suffix)])
	}

	index = tmpl.BitLen/8 - tmpl.HashLen
	if !bytes.Equal(sigMsg[index:index+len(hash)], hash) {
		log.Fatalf("Wrong hash %x", sigMsg[index:index+len(hash)])
	}
}

func BenchmarkSign2048(b *testing.B) {
	signer, _, err := New(testCA)
	if err != nil {
		b.Fatal(err)
	}
	msg := make([]byte, crypto.SHA1.Size())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for {
			_, err = rand.Read(msg)
			if err != nil {
				b.Fatal(err)
			}
			_, err = signer.Sign(nil, msg, crypto.SHA1)
			if _, ok := err.(ErrRetry); ok {
				continue
			}
			if err != nil {
				b.Fatal(err)
			}
			break
		}
	}
}
