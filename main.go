package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const BaseURL = "http://www.omdbapi.com/"

type Film struct {
	Title  string
	Poster string
}

type commandFilmName []string

var countThreads int
var nameFilm commandFilmName

func (i *commandFilmName) String() string {
	return fmt.Sprint(*i)
}

func (i *commandFilmName) Set(value string) error {
	value = strings.Replace(value, ":", "%3A", -1)
	value = strings.Replace(value, "_", "+", -1)
	for _, film := range strings.Split(value, ",") {
		*i = append(*i, film)
	}
	return nil
}

func init() {
	flag.Var(&nameFilm, "f", "разделенный запятыми список фильмов вместо пробела _")
	flag.IntVar(&countThreads, "p", 1, "максимальное кол-во одновременно скачивающихся файлов")
}

func main() {
	flag.Parse()
	fmt.Printf("кол-во одновременно скачивающихся файлов: %v\n", countThreads)
	channel := make(chan string, countThreads)
	start := time.Now()
	for _, film := range nameFilm {
		go SaveFilmPoster(film, channel)
		fmt.Println(<-channel)
	}
	fmt.Println("Скачиванние завершено")
	duration := time.Since(start)
	fmt.Println(duration)

}

func SaveFilmPoster(Namefilm string, outC chan<- string) {
	film := Film{Title: Namefilm}
	err := SearchFilmPoster(&film)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	filename := "./posters/" + film.Title + ".png"
	out, err := os.Create(filename)
	if err != nil {
		fmt.Printf("%v не выполненно ошибка %v\n", film.Title, err)
		return
	}
	defer out.Close()
	resp, err := http.Get(film.Poster)
	if err != nil {
		fmt.Printf("%v не выполненно ошибка %v\n", film.Title, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("%v не выполненно ошибка %v\n", film.Title, err)
		return
	}
	io.Copy(out, resp.Body)
	outC <- "Выполнено " + film.Title
}

func SearchFilmPoster(result *Film) error {
	resp, err := http.Get(BaseURL + "?t=" + result.Title + "&plot=full&apikey=c4653c50")
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return fmt.Errorf("cбой запроса : %s", resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		resp.Body.Close()
		return err
	}
	resp.Body.Close()
	return nil
}
