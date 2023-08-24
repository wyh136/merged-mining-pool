package block

import (
	"encoding/hex"
)

/*******************************************
*******  Merkle Illustration ***************
********************************************

S/O - https://developer.bitcoin.org/reference/block_chain.html#merkle-trees

a   b c   d	e	f <- level1
 \ /   \ /   \ /
  G     H     I   G  <- double up if odd
   \   /       \ /
     J          K <- level3,
	             ... level4,
			     ... levelN
      \___   __/
	      \ /
           L

	Steps is G, J, L

********************************************/

func getMerkleRoot(transactionIDs, steps []string) (string, error) {
	l := len(transactionIDs)

	if l == 0 {
		var empty []byte
		slice := doubleSha256(empty)
		return hex.EncodeToString(slice[:]), nil
	} else if l%2 == 1 {
		transactionIDs = append(transactionIDs, transactionIDs[l-1]) // Last or first?
		l++
	}

	if l == 2 {
		mergedHex, err := mergeHex(transactionIDs[0], transactionIDs[1])
		steps = append(steps, mergedHex)
		return mergedHex, err
	}

	level := transactionIDs
	for l > 1 {
		level, err := scanMerkleLevel(level, steps)
		if err != nil {
			return "", err
		}
		l = len(level)
	}

	return level[0], nil
}

func scanMerkleLevel(pairs, steps []string) ([]string, error) {
	var level []string
	l := len(pairs)
	for i := 0; i < l; i = i + 2 {
		merged, err := mergeHex(pairs[i], pairs[i+1])
		if err != nil {
			return level, err
		}
		if i == 0 {
			steps = append(steps, merged)
		}
		level = append(level, merged)
	}
	return level, nil
}

func mergeHex(one, two string) (string, error) {
	oneBytes, err := hex.DecodeString(one)
	if err != nil {
		return "", err
	}
	twoBytes, err := hex.DecodeString(two)
	if err != nil {
		return "", err
	}

	merged := doubleSha256(append(oneBytes, twoBytes...))

	mergedHex := hex.EncodeToString(merged[:])
	return mergedHex, nil
}

func makeRequestMerkleBranchSteps(transactionIDs []string) ([]string, error) {
	steps := []string{}
	l := len(transactionIDs)

	if l == 0 {
		return steps, nil
	}

	_, err := getMerkleRoot(transactionIDs, steps)
	if err != nil {
		return steps, err
	}

	return steps, nil
}

func makeHeaderMerkleRoot(coinbase []byte, merkleBranchSteps []string) ([]byte, error) {
	block := coinbase
	for _, branch := range merkleBranchSteps {
		branchBytes, err := hex.DecodeString(branch)
		if err != nil {
			return nil, err
		}
		joined := doubleSha256(append(block, branchBytes...))
		block = joined[:]
	}

	return block, nil
}
