package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var Root string

// Chunk is a list of Items
type Chunk struct {
	items map[string]*Item
	mutex sync.Mutex
}

// WordMap is a map stores uniqe word(s)
type WordMap map[string]struct{}

// Item is a representation of cached file
type Item struct {
	id     string
	name   string
	readed bool
	words  WordMap
}

type App struct {
	name  string
	chunk *Chunk
}

// MatchedItem is struct for seach result
type MatchedItem struct {
	id    string
	name  string
	value float64
	score int
}

func (c *Chunk) Push(item *Item) {
	c.items[item.id] = item
}

func (c *Chunk) Remove(item *Item) {
	delete(c.items, item.id)
}

func (c *Chunk) ReadItems() {
	var wg sync.WaitGroup

	// concurrently reading with limit
	limiter := make(chan bool, 10)
	for _, item := range c.items {
		if !item.readed {
			wg.Add(1)
			limiter <- true
			go func(item *Item) {
				defer wg.Done()
				item.Read()
				<-limiter
			}(item)
		}
	}

	wg.Wait()
}

func (c *Chunk) Search(keywords WordMap, max int) []*MatchedItem {
	var nwords int = len(keywords)
	var scoreTable []*MatchedItem

	for _, item := range c.items {
		count := 0
		for w, _ := range keywords {
			if item.IsMatchWord(w) {
				count++
			}
		}

		var val float64 = float64(count) / float64(nwords)
		if val > 0 {
			scoreTable = append(scoreTable, &MatchedItem{
				id:    item.id,
				name:  item.name,
				value: val,
				score: int(math.Ceil(val * 100)),
			})
		}
	}

	sort.Slice(scoreTable, func(i, j int) bool {
		return scoreTable[i].value > scoreTable[j].value
	})

	n := math.Max(float64(0), float64(max))
	n = math.Min(n, float64(len(scoreTable)))

	return scoreTable[:int(n)]
}

func NewWordMap(content string) WordMap {
	re, _ := regexp.Compile("[0-9a-zA-Z\\-]+")
	match := re.FindAllString(content, -1)

	words := make(WordMap)
	for _, m := range match {
		word := strings.ToLower(m)
		words[word] = struct{}{}
	}

	return words
}

func NewItem(fileName string) *Item {
	return &Item{id: md5(fileName), name: fileName, readed: false}
}

func (item *Item) Read() {
	fpath := filepath.Join(Root, item.name)
	content, err := ioutil.ReadFile(fpath)
	if err != nil {
		log.Println("Read error")
	}

	item.words = NewWordMap(string(content))
	item.readed = true
}

func (item *Item) IsMatchWord(word string) bool {
	word = strings.ToLower(word)
	word = strings.TrimSpace(word)
	if _, ok := item.words[word]; ok {
		return true
	}
	return false
}

func NewApp() *App {
	app := &App{
		name: APP_NAME,
	}
	chunk := &Chunk{}
	chunk.items = make(map[string]*Item)
	app.chunk = chunk
	return app
}

// BuildChunk load files to memory
func (app *App) BuildChunk() {
	files, err := ioutil.ReadDir(Root)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		fname := file.Name()
		if strings.HasSuffix(fname, ".txt") {
			item := NewItem(fname)
			app.chunk.Push(item)
		}
	}

	app.chunk.ReadItems()
}

func (app *App) handle(input string) {
	re, _ := regexp.Compile("^:suggest ([0-9]+) (.*)")

	if strings.HasPrefix(input, ":search ") {
		// search
		app.searchHandler(strings.Replace(input, ":search", "", 1))

	} else if strings.HasPrefix(input, ":add ") {
		// add file
		app.addHandler(strings.Replace(input, ":add", "", 1))

	} else if strings.HasPrefix(input, ":rm ") {
		// rm file
		app.rmHandler(strings.Replace(input, ":rm", "", 1))

	} else if m := re.FindStringSubmatch(input); m != nil {
		// suggest
		limit, _ := strconv.Atoi(m[1])
		app.suggestHandler(m[2], limit)

	} else if strings.HasPrefix(input, ":exit") || strings.HasPrefix(input, ":quit") {
		// quit
		fmt.Println("  - bye!")
		os.Exit(0)

	} else {
		// unknown
		fmt.Println("  - unknown command")
	}
}

func (app *App) addHandler(request string) {
	request = strings.TrimSpace(request)

	for _, f := range strings.Split(request, " ") {
		f = strings.TrimSpace(f)

		if fileExist(filepath.Join(Root, f)) {
			item := NewItem(f)
			app.chunk.Push(item)
		}
	}

	app.chunk.ReadItems()
}

func (app *App) rmHandler(request string) {
	request = strings.TrimSpace(request)

	for _, f := range strings.Split(request, " ") {
		f = strings.TrimSpace(f)

		if fileExist(filepath.Join(Root, f)) {
			item := NewItem(f)
			app.chunk.Remove(item)
		}
	}

	app.chunk.ReadItems()
}

func (app *App) searchHandler(request string) {
	request = strings.TrimSpace(request)

	keywords := NewWordMap(request)
	scoreTable := app.chunk.Search(keywords, 10)

	if len(scoreTable) == 0 {
		fmt.Println("   no matches found")
	} else {
		for _, x := range scoreTable {
			fmt.Printf("   - %s : %d%%\n", x.name, x.score)
		}
	}
}

func (app *App) suggestHandler(request string, limit int) {
	request = strings.TrimSpace(request)

	keywords := NewWordMap(request)
	scoreTable := app.chunk.Search(keywords, limit)

	var n int = 0
	for _, x := range scoreTable {
		if x.score == 100 {
			if cacheItem, ok := app.chunk.items[x.id]; ok {
				for w := range cacheItem.words {
					if n >= limit {
						return
					}
					if _, _ok := keywords[w]; _ok {
						continue
					}
					fmt.Printf("   - '%s' %s\n", request, w)
					n++
				}
			}
		}
	}
}

// Start the interactive shell
func (app *App) Start() {
	fmt.Printf("Welcome to %s, Exit with ctrl-c or type ':exit' or ':quit'.\n", app.name)

	app.BuildChunk()

	fmt.Printf("%d file(s) read in directory %s\n", len(app.chunk.items), Root)
	fmt.Print("> ")

	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		input := strings.ToLower(s.Text())
		app.handle(input)
		fmt.Print("> ")
	}

	if err := s.Err(); err != nil {
		log.Fatal("could not read", err)
	}
}
