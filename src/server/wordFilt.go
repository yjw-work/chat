package main

const (
	cstDefTrieNodeSize = 10
	cstFiltChar        = "*"
)

type TTrieNode struct {
	children map[rune]*TTrieNode
}

func NewTrieNode() *TTrieNode {
	return &TTrieNode{
		children: make(map[rune]*TTrieNode, cstDefTrieNodeSize),
	}
}

type TTrie struct {
	root *TTrieNode
}

func NewTrie() *TTrie {
	this := &TTrie{
		root: NewTrieNode(),
	}
	return this
}

func (this *TTrie) Insert(word string) {
	node := this.root
	for _, w := range word {
		if node.children[w] == nil {
			newNode := NewTrieNode()
			node.children[w] = newNode
		}
		node = node.children[w]
	}
}

func (this *TTrie) Filt(content string) string {
	var length int
	for i := 0; i < len(content); {
		length = this.filtPart(content[i:])
		if length <= 0 {
			i++
		} else {
			for j := 0; j < length; j++ {
				content = content[:i] + cstFiltChar + content[i+1:]
				i++
			}
		}
	}
	return content
}

func (this *TTrie) filtPart(part string) int {
	node := this.root
	length := 0
	for _, w := range part {
		if len(node.children) <= 0 {
			break
		}

		if c, ok := node.children[w]; ok {
			node = c
			length++
		} else {
			return 0
		}
	}

	if len(node.children) > 0 {
		return 0
	}

	return length
}
