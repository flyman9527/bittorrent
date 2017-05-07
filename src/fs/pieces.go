package fs

import(
    "io/ioutil"
    "math"
    "crypto/sha1"
    "encoding/base64"
)

const BlockSize int = 16384
type Block []byte

type Piece struct {
    Blocks []Block
}

func (piece *Piece) Hash() string {
    allBytes := []byte{}
    for _, block := range piece.Blocks {
        allBytes = append(allBytes, block...)
    }
    hash := make([]byte, 20)
    actualHash := sha1.Sum(allBytes)

    for i := 0; i < 20; i++ {
        hash[i] = actualHash[i]
    }
    return base64.URLEncoding.EncodeToString(hash)
}

func SplitIntoPieces(path string, pieceLen int) []Piece {
    fileBytes, err := ioutil.ReadFile(path)
    if err != nil {
        panic(err)
    }
    numPieces := NumPieces(pieceLen, len(fileBytes))
    pieces := []Piece{}
    for i:=0; i<numPieces; i++ {
        pieces = append(pieces, getPiece(i, pieceLen, fileBytes))
    }
    return pieces
}

func CombinePieces(path string, pieces []Piece) {
    data := []byte{}
    for _, piece := range pieces {
        for _, block := range piece.Blocks {
            data = append(data, block...)
        }
    }
    err := ioutil.WriteFile(path, data, 0644)
    if err != nil {
        panic(err)
    }
}

func NumBlocksInPiece(piece int, pieceLen int, totalLen int) int {
    // returns the number of blocks for a given piece
    var actualLen int
    numPieces := NumPieces(pieceLen, totalLen)
    if piece == numPieces - 1 && totalLen % BlockSize != 0 {
        // the last piece is irregular
        actualLen = totalLen % BlockSize
    } else {
        actualLen = pieceLen
    }
    return int(math.Ceil(float64(actualLen / BlockSize)))
}

func NumPieces(pieceLen int, totalLen int) int {
    return int(math.Ceil(float64(totalLen) / float64(pieceLen)))
}

func getSubArray(start int, length int, arr []byte) []byte {
    return arr[start : min(start+length, len(arr))]
}

func getPiece(num int, pieceLen int, arr []byte) Piece {
    pieceBytes := getSubArray(num * pieceLen, pieceLen, arr)
    numBlocks := NumBlocksInPiece(num, pieceLen, len(arr))
    blocks := []Block{}
    for i:=0; i<numBlocks; i++ {
        blocks = append(blocks, getBlock(i, pieceBytes))
    }
    return Piece{blocks}
}

func getBlock(block int, piece []byte) Block {
    start := block * BlockSize
    return Block(getSubArray(start, BlockSize, piece))
}

func min(a, b int) int {
    if a <= b {
        return a
    }
    return b
}