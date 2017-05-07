package main

import (
    "testing"
    "util"
    "os"
    "fs"
    "client"
    "tracker"
    "bytes"
    "encoding/gob"
)

func loadDataFromPersister(ps *btclient.Persister) btclient.BTClient {
    data := ps.ReadState()
    cl := btclient.BTClient{}

    if data == nil || len(data) < 1 { // bootstrap without any state?
        return cl
    }
    r := bytes.NewBuffer(data)
    d := gob.NewDecoder(r)
    d.Decode(&cl.Pieces)
    d.Decode(&cl.PieceBitmap)
    return cl
}

func TestTwoClients(t *testing.T) {
    file := "puppy.jpg.torrent"
    seed := "puppy.jpg"

    seederPersister := btclient.MakePersister("test1")
    downloaderPersister := btclient.MakePersister("test2")


    tr := bttracker.StartBTTracker(file, 8000)
    seeder := btclient.StartBTClient("localhost", 6666, file, seed, seederPersister)
    downloader := btclient.StartBTClient("localhost", 6667, file, "", downloaderPersister)

    util.Wait(5000)

    tr.Kill()
    seeder.Kill()
    downloader.Kill()

    res := loadDataFromPersister(downloaderPersister)
    metadata := fs.Read(file)

    if len(res.Pieces) != len(metadata.PieceHashes) {
        util.EPrintf("Client has %d pieces but expected %d pieces\n", len(res.Pieces), len(metadata.PieceHashes))
        t.Fail()
        return
    }
    for i, hash := range metadata.PieceHashes {
        if res.Pieces[i].Hash() != hash {
            util.EPrintf("Piece did not hash correctly\n")
            t.Fail()
            return
        }
    }


    os.Remove(seederPersister.Path)
    os.Remove(downloaderPersister.Path)

    util.EndTest()
}