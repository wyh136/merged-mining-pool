package pool

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"

	"designs.capital/dogepool/bitcoin"
)

func (p *PoolServer) fetchRpcBlockTemplatesAndCacheWork() error {
	var block *bitcoin.BitcoinBlock
	var err error
	template, auxblock, err := p.fetchAllBlockTemplatesFromRPC()
	if err != nil {
		return err
	}

	auxillary := p.config.BlockSignature
	if auxblock != nil {
		mergedPOW := bitcoin.MergedMiningHeader +
			auxblock.Hash + bitcoin.MergedMiningTrailer
		auxillary = auxillary + hexStringToByteString(mergedPOW)
		auxillary = hexStringToByteString(mergedPOW)

		p.templates.AuxBlocks = []bitcoin.AuxBlock{*auxblock}
	}

	primaryName := p.config.GetPrimary()
	rewardPubScriptKey := p.activeNodes[primaryName].RewardPubScriptKey
	extranonceByteReservationLength := 8

	block, p.workCache, err = bitcoin.GenerateWork(&template, auxblock,
		primaryName, auxillary, rewardPubScriptKey,
		extranonceByteReservationLength)
	if err != nil {
		log.Print(err)
	}

	p.templates.BitcoinBlock = *block

	return nil
}

// Main OUTPUT
func (p *PoolServer) recieveWorkFromClient(share bitcoin.Work, client *stratumClient) error {
	primaryBlockTemplate := p.templates.GetPrimary()
	if primaryBlockTemplate.Template == nil {
		return errors.New("Primary block template not yet set")
	}
	auxBlock := p.templates.GetAux1()
	var err error

	primaryBlockHeight := primaryBlockTemplate.Template.Height
	nonce := share[primaryBlockTemplate.NonceSubmissionSlot()].(string)
	extranonce2Slot, _ := primaryBlockTemplate.Extranonce2SubmissionSlot()
	extranonce2 := share[extranonce2Slot].(string)
	nonceTime := share[primaryBlockTemplate.NonceTimeSubmissionSlot()].(string)

	extranonce := client.extranonce1 + extranonce2

	_, err = primaryBlockTemplate.Header(extranonce, nonce, nonceTime)

	if err != nil {
		return err
	}

	heightMessage := fmt.Sprintf("%v", primaryBlockHeight)

	status := verifyShare(&primaryBlockTemplate, auxBlock, share, p.config.PoolDifficulty)

	if status == shareInvalid {
		m := "Invalid share for block %v from %v"
		m = fmt.Sprintf(m, heightMessage, client.ip)
		return errors.New(m)
	}

	m := "Valid share for block %v from %v"
	m = fmt.Sprintf(m, heightMessage, client.ip)
	log.Println(m)

	if status == shareValid {
		return nil
	}

	statusReadable := statusMap[status]

	m = "%v block candidate for block %v from %v"
	m = fmt.Sprintf(m, statusReadable, heightMessage, client.ip)
	log.Println(m)

	err = p.submitAuxBlock(primaryBlockTemplate, *auxBlock, p.config.GetAux1())
	auxStatus := 0
	if err != nil {
		log.Println(err)
		auxStatus = 2
	} else {
		heightMessage = fmt.Sprintf("%v,%v", primaryBlockHeight, auxBlock.Height)
	}

	if status == dualCandidate {
		err = p.submitBlockToChain(primaryBlockTemplate, share, p.config.GetPrimary())
		if err != nil {
			return err
		}
	}

	statusReadable = statusMap[status-auxStatus]

	log.Printf("✅  Successful %v submission of block %v from: %v", statusReadable, heightMessage, client.ip)

	return nil
}

func (pool *PoolServer) generateWorkFromCache(refresh bool) (bitcoin.Work, error) {
	work := append(pool.workCache, interface{}(refresh))

	// TODO - I need to get lower of two bits..

	return work, nil
}

func hexStringToByteString(hexStr string) string {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}
