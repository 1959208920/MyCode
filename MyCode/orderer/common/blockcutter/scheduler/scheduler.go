package scheduler

import (
	"time"

	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	utils "github.com/hyperledger/fabric/protoutil"
)

var logger = flogging.MustGetLogger("orderer.common.blockcutter.scheduler")

func ScheduleTxn(batch []*cb.Envelope) []*cb.Envelope {

	initStrcut(len(batch))
	unMarshalAndSort(batch)

	if len(batch) < 2 {
		return batch
	}

	newbatch := make([]*cb.Envelope, 0)


	// reorder
	schedule, subs := reorderBatch()
	logger.Info("subs:", subs)


	
	// build new batch
	for i, txnID := range schedule {
		logger.Info("schedule ordering: ", i, txnID)
		newbatch = append(newbatch, pendingBatch[txnID])
	}

	return newbatch
}

func initStrcut(size int) {
	pendingBatch = make(map[int]*cb.Envelope)
	scheduler = NewTxnScheduler(uint32(size))
}

func unMarshalAndSort(batch []*cb.Envelope) {

	for i, msg := range batch {

		resppayload, err := utils.GetActionFromEnvelopeMsg(msg)
		if err != nil {
			logger.Info("err 1")
		}
		txRWSet := &rwsetutil.TxRwSet{}
		err = txRWSet.FromProtoBytes(resppayload.Results)
		if err != nil {
			logger.Info("err 2")
		}

		ns := txRWSet.NsRwSets[1]
		readSet := make([]uint64, maxUniqueKeys/64)
		writeSet := make([]uint64, maxUniqueKeys/64)
		tid := int32(len(scheduler.pendingTxns))

		// reorder
		readKeys := []string{}
		writeKeys := []string{}
		defer func(start time.Time) {
			elapsed := time.Since(start).Nanoseconds() / 1000
			logger.Infof("Process txn with read keys %v and write keys %v in %d us", readKeys, writeKeys, elapsed)
		}(time.Now())

		for _, write := range ns.KvRwSet.Writes {
			if writeKey := write.GetKey(); validKey(writeKey) {
				writeKeys = append(writeKeys, writeKey)

				key, ok := scheduler.uniqueKeyMap[writeKey]

				if !ok {
					// if the key is not found, insert and increment
					// the key counter
					scheduler.uniqueKeyMap[writeKey] = scheduler.uniqueKeyCounter
					key = scheduler.uniqueKeyCounter
					scheduler.uniqueKeyCounter += 1
				}

				// set the respective bit in the writeSet
				index := key / 64
				writeSet[index] |= (uint64(1) << (key % 64))
			}
		}

		for _, read := range ns.KvRwSet.Reads {
			if readKey := read.GetKey(); validKey(readKey) {
				readVer := read.GetVersion()
				readKeys = append(readKeys, readKey)

				key, ok := scheduler.uniqueKeyMap[readKey]
				if !ok {
					// if the key is not found, it is inserted. So increment
					// the key counter
					scheduler.uniqueKeyMap[readKey] = scheduler.uniqueKeyCounter
					key = scheduler.uniqueKeyCounter
					scheduler.uniqueKeyCounter += 1
				}

				ver, ok := scheduler.keyVersionMap[key]
				if ok {
					if ver.BlockNum == readVer.BlockNum && ver.TxNum == readVer.TxNum {
						scheduler.keyTxMap[key] = append(scheduler.keyTxMap[key], tid)
					} else {
						// It seems to abort the previous txns with for the unmatched version
						// logger.Infof("Invalidate txn %v", r.keyTxMap[key])
						for _, tx := range scheduler.keyTxMap[key] {
							scheduler.invalid[tx] = true
						}
						scheduler.keyTxMap[key] = nil
					}
				} else {
					scheduler.keyTxMap[key] = append(scheduler.keyTxMap[key], tid)
					scheduler.keyVersionMap[key] = readVer
				}

				index := key / 64
				readSet[index] |= (uint64(1) << (key % 64))
			}
		}

		scheduler.txReadSet[tid] = readSet
		scheduler.txWriteSet[tid] = writeSet
		scheduler.pendingTxns = append(scheduler.pendingTxns, i)
		pendingBatch[i] = msg
	}
}
