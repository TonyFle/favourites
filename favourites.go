package main

import (
	"bytes"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"
)

var modeType string = ""
var b, ss bytes.Buffer
var idType string = ""

type PageVariables struct {
	Data   string
	Mode   string
	Id     string
	Search string
}

type data struct {
	fav_id   string
	category string
	url      string
	link     string
	target   string
}

func empty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

func pickCat(newCat, chgCat, cat string) string {
	if len(strings.TrimSpace(newCat)) != 0 {
		return newCat
	} else if len(strings.TrimSpace(chgCat)) != 0 {
		return chgCat
	} else {
		return cat
	}
}

func getTarget(target string) string {
	if len(strings.TrimSpace(target)) != 0 {
		return target
	} else {
		return "_blank"
	}
}

func (d data) sqlUpdate() {


	db, err := sql.Open("sqlite3", "./fav.db")
	checkErr(err)

	stmt, err := db.Prepare("UPDATE favourites set category = ?, url = ?, link = ?, target = ? where fav_id = ?")
	checkErr(err)

	_, err = stmt.Exec(d.category, d.url, d.link, d.target, d.fav_id)
	checkErr(err)

	db.Close()

}

func (d data) sqlInsert() {

	if empty(d.category) || empty(d.url) || empty(d.category) {
		return //data missing so nothing to do
	}

	var count int

	db, err := sql.Open("sqlite3", "./fav.db")
	checkErr(err)

	err = db.QueryRow("SELECT count(*) as count  from favourites where url=? and category =?", d.url, d.category).Scan(&count)
	checkErr(err)
	if count == 0 {
		stmt, err := db.Prepare("INSERT INTO favourites (category, url, link, target) VALUES (?,?,?,?)")
		checkErr(err)

		_, err = stmt.Exec(d.category, d.url, d.link, d.target)
		checkErr(err)
	}

	db.Close()

}

func (d data) sqlDelete() {

	db, err := sql.Open("sqlite3", "./fav.db")
	checkErr(err)

	stmt, err := db.Prepare("DELETE FROM favourites where fav_id =?")
	checkErr(err)

	_, err = stmt.Exec(d.fav_id)
	checkErr(err)

	db.Close()

}

func urlsearch(keywords string) {
	ss.Reset()
	db, err := sql.Open("sqlite3", "./fav.db")
	checkErr(err)
	var fav_id string
	var category string
	var target string
	var url string
	var link string
	var lastcat string = "none"

	rows, err := db.Query("SELECT * FROM favourites where (url like '%" + keywords + "%' or  link like '%" + keywords + "%') order by category ASC, link ASC")
	checkErr(err)

	for rows.Next() {
		err := rows.Scan(&fav_id, &category, &url, &link, &target)
		checkErr(err)

		if category != lastcat {
			ss.WriteString(string("<tr><td><br>" + category))
			lastcat = category
		} //end if
		ss.WriteString(string("<tr><td><a href=\"" + url + "\" target=\"" + target + "\">" + link + "</a></td></tr>\n"))

	} //end for rows

	rows.Close() //good habit to close

	db.Close()

} //end func

func editPage(favId string) {

	b.Reset()
	db, err := sql.Open("sqlite3", "./fav.db")
	checkErr(err)
	var fav_id string
	var category string
	var url string
	var link string
	var target string
	var cat_list string
	err = db.QueryRow("SELECT * FROM favourites where fav_id ="+favId).Scan(&fav_id, &category, &url, &link, &target)
	checkErr(err)
	rows, err := db.Query("SELECT DISTINCT(category) FROM favourites ORDER BY category")
	checkErr(err)

	b.WriteString(string("<form action=\"fav\" method=post>\n"))
	b.WriteString(string("<input type=\"hidden\" name=\"domod\" value=1>\n"))
	b.WriteString(string("<input type=\"hidden\" name=\"fav_id\" value=\"" + fav_id + "\">\n"))
	b.WriteString(string("<input type=\"hidden\" name=\"category\" value=\"" + category + "\">\n"))
	b.WriteString(string("<table>"))
	b.WriteString(string("<tr><td>Current category</td><td>&nbsp; " + category + "</td></tr>\n"))
	b.WriteString(string("<tr><td>Amend category</td><td><select name=\"chg_cat\">\n"))
	b.WriteString(string("<option value=></option>\n"))

	for rows.Next() {
		err = rows.Scan(&cat_list)
		checkErr(err)
		b.WriteString(string("<option value='" + cat_list + "'> " + cat_list + "\n"))
	}

	b.WriteString(string("</select>\n"))
	b.WriteString(string("<tr><td>New category</td><td><input type=\"text\" size=60 name=\"new_cat\"></td></tr>\n"))
	b.WriteString(string("<tr><td>URL</td><td><input type=\"text\" size=60 name=\"url\" value=\"" + url + "\"></td></tr>\n"))
	b.WriteString(string("<tr><td>Link text</td><td><input type=\"text\" size=60 name=\"link\" value=\"" + link + "\"></td></tr>\n"))
	b.WriteString(string("<tr><td>Target</td><td><input type=\"text\" size=60 name=\"target\" value=\"" + target + "\"></td></tr>\n"))
	b.WriteString(string("<tr><td colspan=2><input type=submit value=Modify></td></tr>"))
	b.WriteString(string("</table>"))
	b.WriteString(string("</form><br>\n"))

	b.WriteString(string("<br><form action=\"fav\" method=post>\n"))
	b.WriteString(string("<input type=\"hidden\" name=\"dodel\" value=1>\n"))
	b.WriteString(string("<input type=\"hidden\" name=\"fav_id\" value=\"" + fav_id + "\">\n"))
	b.WriteString(string("Category: " + category + "<br>\n"))
	b.WriteString(string("URL: " + url + "<br>\n"))
	b.WriteString(string("Link: " + link + "<br>\n"))
	b.WriteString(string("<input type=submit value=Delete>"))
	b.WriteString(string("</form><br>"))

	rows.Close() //good habit to close

	db.Close()

}

func insertPage() {

	b.Reset()
	db, err := sql.Open("sqlite3", "./fav.db")
	checkErr(err)
	rows, err := db.Query("SELECT DISTINCT(category) FROM favourites ORDER BY category")
	checkErr(err)
	var category string

	b.WriteString(string("<form action=\"fav\" method=post>\n"))
	b.WriteString(string("<input type=\"hidden\" name=\"doins\" value=1>\n"))
	b.WriteString(string("<table>"))
	b.WriteString(string("<tr>"))
	b.WriteString(string("<td valign=top>Category</td>"))
	b.WriteString(string("<td>"))
	b.WriteString(string("<select name=\"category\">\n"))
	b.WriteString(string("<option value=></option>\n"))

	for rows.Next() {
		err = rows.Scan(&category)
		checkErr(err)
		b.WriteString(string("<option value='" + category + "'> " + category + "\n"))
	}

	b.WriteString(string("</select>\n"))
	b.WriteString(string("<br><input type=\"text\" size=60 name=\"new_cat\">"))
	b.WriteString(string("</td></tr>\n"))
	b.WriteString(string("<tr><td>URL</td><td><input type=\"text\" size=60 name=\"url\"></td></tr>\n"))
	b.WriteString(string("<tr><td>Link</td><td><input type=\"text\" size=60 name=\"link\"></td></tr>\n"))
	b.WriteString(string("<tr><td>Target</td><td><input type=\"text\" size=60 name=\"target\"></td></tr>\n"))
	b.WriteString(string("<tr><td colspan=2><input type=submit value=Insert></td></tr>"))
	b.WriteString(string("</table>"))
	b.WriteString(string("</form><br>\n"))

	rows.Close() //good habit to close

	db.Close()

}

func normalAndEditPage() {

	b.Reset()
	db, err := sql.Open("sqlite3", "./fav.db")
	checkErr(err)
	rows, err := db.Query("SELECT * FROM favourites order by category ASC, link ASC")
	checkErr(err)
	var fav_id string
	var category string
	var url string
	var link string
	var target string
	oldCategory := ""
	menuNum := 0
	firstTime := true

	for rows.Next() {
		err = rows.Scan(&fav_id, &category, &url, &link, &target)
		checkErr(err)
		if oldCategory != category {
			if firstTime {
				firstTime = false
			} else {
				b.WriteString(string("</dd></dt>\n"))
			} //end
			menuNum++
			b.WriteString(string("<dt onclick=\"javascript:show('lmenu" + strconv.Itoa(menuNum) + "');\"><img src=\"/images/folder.gif\" width=18 height=16> " + category + "\n"))
			b.WriteString(string("<dd id=lmenu" + strconv.Itoa(menuNum) + ">\n"))
		} //end if
		oldCategory = category

		if modeType == "Edit Mode" {
			b.WriteString(string("<table><tr><td width=300><a href=\"" + url + "\" target=\"" + target + "\">" + link + "</a></td><td><a href=\"fav?fav_id=" + fav_id + "&mod=1\">Edit</a></td></tr></table>\n"))
		} else {
			b.WriteString(string("<table><tr><td width=300><a href=\"" + url + "\" target=\"" + target + "\">" + link + "</a></td></tr></table>\n"))
		}

	} // end for

	rows.Close() //good habit to close

	b.WriteString(string("</dl>\n"))

	db.Close()

}

func favourites(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	if len(r.Form) == 0 {
		modeType = ""
		normalAndEditPage()
	}

	if _, exist := r.Form["normal"]; exist {
		modeType = ""
		normalAndEditPage()
	}

	if _, exist := r.Form["insert"]; exist {
		modeType = "Insert Mode"
		insertPage()
	}

	if _, exist := r.Form["doins"]; exist {
		cat := r.Form["category"]
		new_cat := r.Form["new_cat"]
		url := r.Form["url"]
		link := r.Form["link"]
		targ := r.Form["target"]
		category := pickCat(new_cat[0], cat[0], "")
		target := getTarget(targ[0])
		d := data{category: category, url: url[0], link: link[0], target: target}
		d.sqlInsert()
		modeType = ""
		normalAndEditPage()
	}

	if _, exist := r.Form["edit"]; exist {
		modeType = "Edit Mode"
		normalAndEditPage()
	}

	if _, exist := r.Form["dodel"]; exist {
		if id, exist := r.Form["fav_id"]; exist {
			d := data{fav_id: id[0]}
			d.sqlDelete()
		}
		modeType = ""
		normalAndEditPage()
	}

	if _, exist := r.Form["urlsearch"]; exist {
		modeType = ""
		normalAndEditPage()
		keyword, _ := r.Form["keyword"]
		urlsearch(keyword[0])
	}

	if _, exist := r.Form["mod"]; exist {
		modeType = "Edit Mode"
		favId, _ := r.Form["fav_id"]
		editPage(favId[0])
	}

	if _, exist := r.Form["domod"]; exist {
		id, _ := r.Form["fav_id"]
		orig_cat := r.Form["category"]
		chg_cat := r.Form["chg_cat"]
		new_cat := r.Form["new_cat"]
		url := r.Form["url"]
		link := r.Form["link"]
		targ := r.Form["target"]
		category := pickCat(new_cat[0], chg_cat[0], orig_cat[0])
		target := getTarget(targ[0])
		d := data{fav_id: id[0], category: category, url: url[0], link: link[0], target: target}
		d.sqlUpdate()
		modeType = ""
		normalAndEditPage()
	}

	if _, exist := r.Form["google"]; exist {
		keyword, _ := r.Form["keyword"]
		kw := url.QueryEscape(keyword[0])
		w.Header().Set("Location", "http://www.google.co.uk/#sclient=psy&hl=en&q="+kw)
		w.WriteHeader(301)
	}

	if _, exist := r.Form["eBay"]; exist {
		keyword, _ := r.Form["keyword"]
		kw := url.QueryEscape(keyword[0])
		w.Header().Set("Location", "http://www.ebay.co.uk/sch/i.html?_nkw="+kw)
		w.WriteHeader(301)
	}

	if _, exist := r.Form["youtube"]; exist {
		keyword, _ := r.Form["keyword"]
		kw := url.QueryEscape(keyword[0])
		w.Header().Set("Location", "http://www.youtube.com/results?search_query="+kw+"&aq=f")
		w.WriteHeader(301)
	}

	if _, exist := r.Form["image"]; exist {
		keyword, _ := r.Form["keyword"]
		kw := url.QueryEscape(keyword[0])
		w.Header().Set("Location", "http://images.google.co.uk/images?hl=en&q="+kw)
		w.WriteHeader(301)
	}

	if _, exist := r.Form["imdb"]; exist {
		keyword, _ := r.Form["keyword"]
		kw := url.QueryEscape(keyword[0])
		w.Header().Set("Location", "http://www.imdb.com/find?q="+kw+"&s=all")
		w.WriteHeader(301)
	}

	if _, exist := r.Form["shopping"]; exist {
		keyword, _ := r.Form["keyword"]
		kw := url.QueryEscape(keyword[0])
		w.Header().Set("Location", "http://www.google.co.uk/products?hl=en&q="+kw)
		w.WriteHeader(301)
	}

	if _, exist := r.Form["maps"]; exist {
		keyword, _ := r.Form["keyword"]
		kw := url.QueryEscape(keyword[0])
		w.Header().Set("Location", "http://maps.google.co.uk/maps?hl=en&q="+kw)
		w.WriteHeader(301)
	}

	if _, exist := r.Form["wiki"]; exist {
		keyword, _ := r.Form["keyword"]
		kw := url.QueryEscape(keyword[0])
		w.Header().Set("Location", " http://en.wikipedia.org/w/index.php?search="+kw)
		w.WriteHeader(301)
	}

	if _, exist := r.Form["amazon"]; exist {
		keyword, _ := r.Form["keyword"]
		kw := url.QueryEscape(keyword[0])
		w.Header().Set("Location", " https://www.amazon.co.uk/s/ref=nb_sb_noss_2?url=search-alias%3Daps&field-keywords="+kw)
		w.WriteHeader(301)
	}

	if _, exist := r.Form["db2v101"]; exist {
		keyword, _ := r.Form["keyword"]
		kw := url.QueryEscape(keyword[0])
		w.Header().Set("Location", "  https://www.ibm.com/support/knowledgecenter/search/"+kw+"?scope=SSEPGG_10.1.0")
		w.WriteHeader(301)
	}

	var cookie, err = r.Cookie("id")
	if err == nil {
		idType = cookie.Value
	}

	HomePageVars := PageVariables{
		Data:   b.String(),
		Mode:   modeType,
		Id:     idType,
		Search: ss.String(),
	}

	t, _ := template.ParseFiles("favourites.html")
	t.Execute(w, HomePageVars)

}

func main() {
	http.HandleFunc("/", favourites) // setting router rule
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./images"))))

	err := http.ListenAndServe(":8080", nil) // setting listening port
	//err  := http.ListenAndServeTLS(":8443", "server.crt", "server.key", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
