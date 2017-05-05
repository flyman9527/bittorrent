package btclient

import (
	"btnet"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/gob"
	"net"
	"net/http"
	"time"
	"util"
	"strconv"
	"errors"
	"io/ioutil"
)

type Block []byte

type Piece struct {
	blocks []Block
}

func peerTimeout() time.Duration {
  return time.Millisecond * 2000
}

const BlockSize int = 16384
const DialTimeout = time.Millisecond * 100

func (cl *BTClient) startServer() {
	btnet.StartTCPServer(cl.ip+":"+cl.port, cl.messageHandler)
}

type requestParams struct {
	peerId     string
	ip         string
	port       string
	uploaded   int
	downloaded int
	left       int
	infoHash   string
}

func sendRequest(addr string, req *requestParams) ([]byte, error) {
	url := addr + "/?peer_id=" + req.peerId +
		"&port=" + req.port + "&ip=" + req.ip + "&uploaded=" +
		strconv.Itoa(req.uploaded) + "&downloaded=" + strconv.Itoa(req.downloaded) +
		"&left=" + strconv.Itoa(req.left) + "&info_hash=" + req.infoHash
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.New("Error sending request")
	}
	if resp.Status != "200 OK" {
		return nil, errors.New("Wrong response status code")
	}
	bodyBytes, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		return nil, errors.New("Failure reading response body")
	}
	return bodyBytes, nil
}

func (cl *BTClient) contactTracker(url string) {
	// TODO: fill this in
	util.IPrintf("contacting tracker at %s", url)
	res, err := sendRequest(url, &requestParams{})
	util.IPrintf("res %s", res)
	util.IPrintf("err %v", err)
}

// func (cl *BTClient) listenForPeers() {
//   // Set up stuff
//   cl.startServer()
//
//   //TODO: Send initial hello packets
// }

func (cl *BTClient) seed() {
	for {
		if cl.CheckShutdown() {
			return
		}
		cl.mu.Lock()
		go cl.contactTracker(cl.torrent.TrackerUrl)

		//TODO: only here for compilation
		// fmt.Println(file)

		cl.mu.Unlock()
		util.Wait(1000)
	}
}

func (cl *BTClient) connectToPeer(addr string) {
	conn, err := net.DialTimeout("tcp", addr, DialTimeout)
	if err != nil {
		// TODO: try again or mark peer as down
	}
	// Create hello message
	// TODO: here for compilation hehe
	if conn != nil {
	}
	// Send hello message to peer

}

func (cl *BTClient) sendBlock(index int, begin int, length int, peer btnet.Peer) {
	if !cl.PieceBitmap[index] {
		// we don't have this piece yet
		return
	}
	if length != BlockSize {
		// the requester is using a different block size
		// deny the request for simplicity
		return
	}
	if begin%BlockSize != 0 {
		return
	}
	blockIndex := begin / BlockSize
	data := cl.Pieces[index].blocks[blockIndex]
	go cl.sendPieceMessage(peer, index, begin, length, data)
}

func (cl *BTClient) saveBlock(index int, begin int, length int, block []byte) {
	if begin%BlockSize != 0 {
		return
	}
	if length < BlockSize {
		return
	}
	blockIndex := begin / BlockSize
	cl.Pieces[index].blocks[blockIndex] = block[:BlockSize]

	if _, ok := cl.blockBitmap[index]; !ok {
		cl.blockBitmap[index] = make([]bool, cl.numBlocks, cl.numBlocks)
	}
	cl.blockBitmap[index][blockIndex] = true

	if allTrue(cl.blockBitmap[index]) {
		// hash and save piece
		if cl.Pieces[index].hash() != cl.torrent.PieceHashes[index] {
			delete(cl.blockBitmap, index)
			return
		}
		cl.PieceBitmap[index] = true
		cl.persistPieces()
	}
}

func (piece *Piece) hash() string {
	allBytes := []byte{}
	for _, block := range piece.blocks {
		allBytes = append(allBytes, block...)
	}
	hash := make([]byte, 20)
	actualHash := sha1.Sum(allBytes)

	for i := 0; i < 20; i++ {
		hash[i] = actualHash[i]
	}
	return base64.URLEncoding.EncodeToString(hash)
}

func (cl *BTClient) persistPieces() {
	w := new(bytes.Buffer)
	e := gob.NewEncoder(w)
	e.Encode(cl.Pieces)
	e.Encode(cl.PieceBitmap)
	data := w.Bytes()
	cl.persister.SaveState(data)
}

func (cl *BTClient) loadPieces(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	r := bytes.NewBuffer(data)
	d := gob.NewDecoder(r)
	d.Decode(&cl.Pieces)
	d.Decode(&cl.PieceBitmap)
}

// Helper
func allTrue(arr []bool) bool {
	for _, entry := range arr {
		if !entry {
			return false
		}
	}
	return true
}

func (cl *BTClient) sendPieceMessage(peer btnet.Peer, index int, begin int, length int, data []byte) {
	message := btnet.PeerMessage{
		Type:   btnet.Piece,
		Index:  int32(index),
		Begin:  begin,
		Length: length,
		Block:  data}
	util.TPrintf("sending message - %v\n", message)
	// TODO send message
}

func (cl *BTClient) SendPeerMessage(addr *net.TCPAddr, message btnet.PeerMessage) {
  peer, ok := cl.peers[addr.String()]
  if !ok {
    // TODO: Something went wrong
    // Try dialing
    // connection := DoDial(addr, data)
    infoHash := ""
    peerId := ""
    bitfieldLength := 0
    peer := btnet.InitializePeer(addr, infoHash, peerId, bitfieldLength, nil)
    cl.peers[addr.String()] = peer

    // Start go routine that handles the closing of the tcp connection if we dont
    // get a keepAlive signal


    // Separate go routine for sending keepalive signals
    go func() {
      for {
        time.Sleep(peerTimeout() / 2)
        msg := btnet.PeerMessage{KeepAlive: true}
        data := btnet.EncodePeerMessage(msg)

        _,err := peer.Conn.Write(data)
        if (err != nil) {
          // Connection is probably closed
          break
        }
      }
    }()
  }
  // send data
  data := btnet.EncodePeerMessage(message)
  peer.Conn.Write(data)
  // tcpAddr, _ := net.ResolveTCPAddr("tcp", (*addr).String())
  // connection.Close()
}

func (cl *BTClient) messageHandler(conn net.Conn) {
	// Max message size: 2^17 = 131072 (128KB)
	// buf := make([]byte, 131072)
	// bytesRead, err := conn.(*net.TCPConn).Read(buf)
	// fmt.Println(bytesRead)
	// if err != nil {
	//   fmt.Println("hi")
	// 	fmt.Println(err)
	// }
  // Check if this is a new connection
  // If so we need to initialize the Peer
  util.Printf("~~~ Got a connection! ~~~\n")

  peer, ok := cl.peers[conn.RemoteAddr().String()]
	if !ok {
		util.Printf("Missing peer: %v\n", conn.RemoteAddr().String())
		// InitializePeer
		// TODO: use the actual length len(cl.torrent.PieceHashes)
    // TODO: Get the actual infoHash string and peerId string
    // util.Printf("This is receiving a connection: %v\n", conn.RemoteAddr())
		newPeer :=  btnet.InitializePeer(conn.RemoteAddr().(*net.TCPAddr), "01234567890123456789", "01234567890123456789", 10, conn.(*net.TCPConn))
		util.Printf("newPeer.Addr.String(): " + newPeer.Addr.String()+"\n")
		if len(newPeer.Addr.String()) < 3 {
			conn.(*net.TCPConn).Close()
			util.EPrintf("Dropping peer connection: Bad handshake\n")
			return
		}
		util.Printf("adding newPeer: %v", newPeer.Conn)
		cl.peers[conn.RemoteAddr().String()] = newPeer
		util.Printf("cl.peers-- %v\n", cl.peers)

		go func() {
			for {
				select {
				case <- peer.KeepAlive:
					// Do nothing, this is good
				case <- time.After(peerTimeout()):
					peer.Conn.Close()
					delete(cl.peers, peer.Addr.String())
					// cl.peers[addr] = nil
					break
				}
			}
		}()

    return
  }

	util.Printf("Already have peer: %v, in %v", conn.RemoteAddr().String(), cl.peers)

  // Process the message
	buf := btnet.ReadMessage(conn.(*net.TCPConn))

	peerMessage := btnet.DecodePeerMessage(buf)
	// Massive switch case that would handle incoming messages depending on message type

	// peerMessage := btnet.PeerMessage{}  // empty for now, TODO
	cl.mu.Lock()
	defer cl.mu.Unlock()

  if peerMessage.KeepAlive {
    peer.KeepAlive <- true
  }

	switch peerMessage.Type {
	case btnet.Choke:
		peer.Status.PeerChoking = true
	case btnet.Unchoke:
		peer.Status.PeerChoking = false
	case btnet.Interested:
		peer.Status.PeerInterested = true
	case btnet.NotInterested:
		peer.Status.PeerInterested = false
	case btnet.Have:
		peer.Bitfield[peerMessage.Index] = true
	case btnet.Bitfield:
		peer.Bitfield = peerMessage.Bitfield
	case btnet.Request:
		cl.sendBlock(int(peerMessage.Index), peerMessage.Begin, peerMessage.Length, peer)
	case btnet.Piece:
		cl.saveBlock(int(peerMessage.Index), peerMessage.Begin, peerMessage.Length, peerMessage.Block)
	case btnet.Cancel:
		// TODO
	default:
		// keepalive
		// TODO?
	}
}
