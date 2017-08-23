package posseg

import (
	"math"
	"sync"

	"github.com/jaysharp/jiebago/dictionary"
)

type MyDictionary struct {
	Dictionary

	newFreqMap map[string]float64
	newPosMap  map[string]string
}

// 在原jieba基础上增加新的词典结构
var jiebaCommonDictOnce sync.Once
var _jiebaCommonDict *Dictionary

func CommonDictIns(fileName string) *Dictionary {
	jiebaCommonDictOnce.Do(func() {
		_jiebaCommonDict = &Dictionary{FreqMap: make(map[string]float64), PosMap: make(map[string]string)}
		_jiebaCommonDict.LoadDictionary(fileName)
	})
	return _jiebaCommonDict
}

func ReloadCommonDictIns(fileName string) error {
	_jiebaCommonDict = &Dictionary{FreqMap: make(map[string]float64), PosMap: make(map[string]string)}
	_jiebaCommonDict.LoadDictionary(fileName)

	return nil
}

// Frequency returns the frequency and existence of give word
func (myDict *MyDictionary) Frequency(key string) (float64, bool) {
	myDict.RLock()
	freq, ok := myDict.newFreqMap[key]
	if !ok {
		freq, ok = myDict.FreqMap[key]
	}
	myDict.RUnlock()
	return freq, ok
}

// Pos returns the POS and existence of give word
func (myDict *MyDictionary) Pos(key string) (string, bool) {
	myDict.RLock()
	pos, ok := myDict.newPosMap[key]
	if !ok {
		pos, ok = myDict.PosMap[key]
	}
	myDict.RUnlock()
	return pos, ok
}

// AddToken adds one token
func (myDict *MyDictionary) AddMyToken(token dictionary.Token) {
	myDict.Lock()
	myDict.addToken(token)
	myDict.Unlock()
	myDict.updateLogTotal()
}

func (myDict *MyDictionary) DelMyWord(word string) {
	freq := myDict.newFreqMap[word]
	delete(myDict.newFreqMap, word)
	delete(myDict.newPosMap, word)
	myDict.Total -= freq

	// 自定义名词之间可能共享字符，对于newFreqMap中的一些单字符垃圾暂不作处理
}

func (myDict *MyDictionary) addToken(token dictionary.Token) {
	myDict.newFreqMap[token.Text()] = token.Frequency()
	myDict.Total += token.Frequency()
	runes := []rune(token.Text())
	n := len(runes)
	for i := 0; i < n; i++ {
		frag := string(runes[:i+1])
		if _, ok := myDict.newFreqMap[frag]; !ok {
			myDict.newFreqMap[frag] = 0.0
		}
	}
	if len(token.Pos()) > 0 {
		myDict.newPosMap[token.Text()] = token.Pos()
	}
}

// A Dictionary represents a thread-safe dictionary used for word segmentation.
type Dictionary struct {
	Total, LogTotal float64
	FreqMap         map[string]float64
	PosMap          map[string]string
	sync.RWMutex
}

// Load loads all tokens from given channel
func (d *Dictionary) Load(ch <-chan dictionary.Token) {
	d.Lock()
	for token := range ch {
		d.addToken(token)
	}
	d.Unlock()
	d.updateLogTotal()
}

// AddToken adds one token
func (d *Dictionary) AddToken(token dictionary.Token) {
	d.Lock()
	d.addToken(token)
	d.Unlock()
	d.updateLogTotal()
}

func (d *Dictionary) addToken(token dictionary.Token) {
	d.FreqMap[token.Text()] = token.Frequency()
	d.Total += token.Frequency()
	runes := []rune(token.Text())
	n := len(runes)
	for i := 0; i < n; i++ {
		frag := string(runes[:i+1])
		if _, ok := d.FreqMap[frag]; !ok {
			d.FreqMap[frag] = 0.0
		}
	}
	if len(token.Pos()) > 0 {
		d.PosMap[token.Text()] = token.Pos()
	}
}

func (d *Dictionary) updateLogTotal() {
	d.LogTotal = math.Log(d.Total)
}

// Frequency returns the frequency and existence of give word
func (d *Dictionary) Frequency(key string) (float64, bool) {
	d.RLock()
	freq, ok := d.FreqMap[key]
	d.RUnlock()
	return freq, ok
}

// Pos returns the POS and existence of give word
func (d *Dictionary) Pos(key string) (string, bool) {
	d.RLock()
	pos, ok := d.PosMap[key]
	d.RUnlock()
	return pos, ok
}

func (d *Dictionary) LoadDictionary(fileName string) error {
	return dictionary.LoadDictionary(d, fileName)
}
