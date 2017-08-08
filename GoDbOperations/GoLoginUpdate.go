package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/satori/go.uuid"
)

var dbSession = map[string]string{}

func home(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println("SAYFAACTA")
	template, err := template.ParseFiles("home.html")
	if err != nil {
		w.Write([]byte("500 server error"))
	}
	template.Execute(w, nil)
}

func kayit(response http.ResponseWriter, request *http.Request) {
	//db_username Ve db_password kısımlarını değiştiriniz
	db, err := sql.Open("mysql", "db_username:db_password@/db_name")
	if err != nil {
		fmt.Println("Database acilirken bir sikinti olustu")
	}
	//defer db.Close()
	request.ParseForm()
	kadi := request.Form["kayit_username"][0]
	ksifre := request.Form["kayit_password"][0]
	kemail := request.Form["kayit_email"][0]
	fmt.Println("kayittayiz")
	//execResults, err := db.Exec("INSERT INTO uyeler (username,sifre,email) VALUES (" + kadi + "," + ksifre + "," + kemail + ")")
	execResults, err := db.Exec("INSERT INTO uyeler (username,sifre,email) VALUES (?,?,?)", kadi, ksifre, kemail)
	if err != nil {
		fmt.Println("insert execute edilirken bir sıkıntı oldu")
	} else {
		sonid, _ := execResults.LastInsertId()
		affectedRow, _ := execResults.RowsAffected()
		fmt.Println("son eklenen id: ", sonid, " \n etkilenen satir : ", affectedRow)
	}
	template, err := template.ParseFiles("gologintmp.html")
	if err != nil {
		fmt.Println("template acilamadi")
	}
	str := "Kayit Executeye kadar geldikkk"
	template.Execute(response, str)
}

func login(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", "db_username:db_password@/db_name")
	if err != nil {
		fmt.Println("Database acilirken bir sikinti olustu")
	}
	//defer db.Close()
	r.ParseForm()
	kadi := r.Form["username"][0]
	ksifre := r.Form["password"][0]
	var isAuthenticated bool
	authent_err := db.QueryRow("SELECT IF(COUNT(*),'true','false') FROM uyeler WHERE username = ? AND sifre = ?", kadi, ksifre).Scan(&isAuthenticated)
	if authent_err != nil {
		log.Fatal(err)
	}
	if isAuthenticated == false {
		fmt.Println("isAuthendicated false")
	} else { //!!!!!!!!!!!!!!!!!!!!!!!!!!!SESSIONNN!!!!!!!!!!!!!!!!!!!!!!
		fmt.Println("isAuthendicated true yani isler yolunda")
		if r.FormValue("checkbox") != "" { // Eğer Beni Hatırla'yı check ettiyse kullanıcı tarayıcıyı kapatsa bile login kalabilmesi için COOKIE yollayacağız.
			fmt.Println("Check(if)'deyiz")
			sID := uuid.NewV4()
			c := &http.Cookie{
				Name:  "cookie",
				Value: sID.String(),
			}
			dbSession[c.Value] = kadi // Session'un kendine has Value'si ile Session'u oluşturulan kullanıcı arasında
			//bir ilişki olması için aralarında map oluşturuldu.
			//http.Redirect(w,r,"/sessiOn",http.StatusSeeOther)
			expiration := time.Now().Add(365 * 24 * time.Hour)
			c.Expires = expiration
			http.SetCookie(w, c)
			template, err := template.ParseFiles("gologintmp.html")
			if err != nil {
				fmt.Println("template acilamadi")
			}
			//str := "Login Executeye kadar geldikkk"
			template.Execute(w, c.Name+"1 Yıllık Persistent Cookie Yollandi")
		} else { // Eğer Beni Hatırla check edilmemişse bir SESSION oluşturacağız. Tarayıcı kapanırsa Logout olmuş olacak.
			fmt.Println("Uncheck(else)'deyiz")
			sID := uuid.NewV4()
			c := &http.Cookie{
				Name:  "cookie",
				Value: sID.String(),
			}
			dbSession[c.Value] = kadi // Session'un kendine has Value'si ile Session'u oluşturulan kullanıcı arasında
			//bir ilişki olması için aralarında map oluşturuldu.
			//http.Redirect(w,r,"/sessiOn",http.StatusSeeOther)
			http.SetCookie(w, c)
			template, err := template.ParseFiles("updateadd.html")
			if err != nil {
				fmt.Println("template acilamadi")
			}
			//str := "Login Executeye kadar geldikkk"
			template.Execute(w, "Session Cookie Yollandı")
		}

	}

}

func sessiOn(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println("SAYFAACTA")
	template, err := template.ParseFiles("updateadd.html")
	if err != nil {
		w.Write([]byte("500 server error"))
	}
	template.Execute(w, nil)
}

func update(w http.ResponseWriter, r *http.Request) {
	fmt.Println("~~~~UPDATE'deyiz~~~~~")
	r.ParseForm()
	db, err := sql.Open("mysql", "db_username:db_password@/db_name")
	if err != nil {
		fmt.Println("Database acilirken bir sikinti olustu")
	}
	//defer db.Close()
	caughtCookie, err := r.Cookie("cookie")
	if err != nil {
		fmt.Println("UPDATE cookie errorr")
	}
	eskiKadi := dbSession[caughtCookie.Value] // Kullanıcı bilgilerini değiştirebilmek için eski id'si veya username'ine sahip
	//olmalıyız. Zira 2'si de unique değerlerdir. Burada,gelen Cookie içindeki Value ile önceden oluşturduğumuz map'de
	//eşleşen username'i çekiyoruz.
	kadi := r.Form["degis_username"][0]
	ksifre := r.Form["degis_password"][0]
	kemail := r.Form["degis_email"][0]
	fmt.Println("Degisteyiz")
	//execResults, err := db.Exec("INSERT INTO uyeler (username,sifre,email) VALUES (" + kadi + "," + ksifre + "," + kemail + ")")

	/* // Aşağıdaki kısım,Update için gereksiz imiş,önce bulup sonra değiştirdiği için değişecek olan username'i WHERE'den
	//sonra parametre olarak kullanmak sıkıntılı değil.
		var kid int
		err = db.QueryRow("SELECT id FROM uyeler WHERE username = ?", eskiKadi).Scan(&kid) //Geri kalan herşeyi değişmek için
		//elimizden en azından önceden kalma bir tane veri olmalıdır. O yüzden burada zaten hiç değişmeyen ID'yi çekiyoruz ki
		//UPDATE sorgusunda WHERE kısmına yazabilelim.
		if err != nil {
			fmt.Println("Degistirde Kullanici ismi bulmada sikinti yasandi.")
		} else {
			fmt.Println("UPDATE kullanici id'sine ulastik. id : ", kid)
		}
	*/
	execResults, err := db.Exec("UPDATE uyeler SET username = ? , sifre = ? , email = ? WHERE username = ?", kadi, ksifre, kemail, eskiKadi)
	if err != nil {
		fmt.Println("UPDATE execute edilirken bir sıkıntı oldu")
	} else {
		sID := uuid.NewV4()
		c := &http.Cookie{
			Name:  "cookie",
			Value: sID.String(),
		}
		dbSession[c.Value] = kadi
		http.SetCookie(w, c)
		sonid, _ := execResults.LastInsertId()
		affectedRow, _ := execResults.RowsAffected()
		fmt.Println("UPDATE son eklenen id: ", sonid, " \n etkilenen satir : ", affectedRow)
	}
	template, err := template.ParseFiles("updateadd.html")
	if err != nil {
		fmt.Println("UPDATE template acilamadi")
	}
	str := "UPDATE Executeye kadar geldikkk"
	template.Execute(w, str)
}

func main() {
	http.HandleFunc("/", home)
	http.HandleFunc("/login", login)
	http.HandleFunc("/kayit", kayit)
	http.HandleFunc("/sessiOn", sessiOn)
	http.HandleFunc("/update", update)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe", err)
	}
}
