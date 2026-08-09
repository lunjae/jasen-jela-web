package validation

import (
	"github.com/lunjae/jasen-jela-web/backend/internal/models"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var phoneRE = regexp.MustCompile(`^[+0-9][0-9 ()/-]{5,24}$`)
var slugRE = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var publicIDRE = regexp.MustCompile(`^(?:jasen-jela/)?products/[^/]+/[^/]+$`)

func Product(p models.Product) map[string]string {
	e := map[string]string{}
	required(e, "name", p.Name, 2, 120)
	required(e, "shortDescription", p.ShortDescription, 10, 240)
	required(e, "description", p.Description, 20, 5000)
	required(e, "categoryId", p.CategoryID, 1, 100)
	required(e, "material", p.Material, 2, 100)
	required(e, "color", p.Color, 2, 100)
	validSlug(e, p.Slug)
	if p.Currency != "RSD" && p.Currency != "EUR" {
		e["currency"] = "Valuta mora biti RSD ili EUR."
	}
	if p.Price != nil && *p.Price < 0 {
		e["price"] = "Cena ne može biti negativna."
	}
	if p.Dimensions != nil {
		for key, value := range map[string]*float64{"length": p.Dimensions.Length, "width": p.Dimensions.Width, "height": p.Dimensions.Height} {
			if value != nil && *value < 0 {
				e["dimensions."+key] = "Dimenzija ne može biti negativna."
			}
		}
	}
	primaryCount := 0
	publicIDs := map[string]bool{}
	for index, image := range p.Images {
		key := "images[" + strconv.Itoa(index) + "]"
		u, parseErr := url.ParseRequestURI(image.URL)
		if parseErr != nil || u.Scheme != "https" || u.Host == "" || !publicIDRE.MatchString(image.PublicID) || image.Order != index || len(strings.TrimSpace(image.Alt)) > 200 || publicIDs[image.PublicID] {
			e[key] = "Metapodaci slike nisu ispravni."
		}
		publicIDs[image.PublicID] = true
		if image.IsPrimary {
			primaryCount++
		}
	}
	if len(p.Images) > 0 && primaryCount != 1 {
		e["images"] = "Tačno jedna slika mora biti glavna."
	}
	return e
}
func Category(c models.Category) map[string]string {
	e := map[string]string{}
	required(e, "name", c.Name, 2, 100)
	validSlug(e, c.Slug)
	if len(c.Description) > 1000 {
		e["description"] = "Opis je predugačak."
	}
	return e
}
func validSlug(e map[string]string, slug string) {
	if !slugRE.MatchString(slug) || len(slug) > 140 {
		e["slug"] = "Slug nije ispravan."
	}
}
func Inquiry(i models.Inquiry) map[string]string {
	e := map[string]string{}
	required(e, "fullName", i.FullName, 2, 120)
	if _, x := mail.ParseAddress(strings.TrimSpace(i.Email)); x != nil {
		e["email"] = "Unesite ispravnu email adresu."
	}
	if !phoneRE.MatchString(strings.TrimSpace(i.Phone)) {
		e["phone"] = "Unesite ispravan broj telefona."
	}
	required(e, "message", i.Message, 10, 3000)
	return e
}
func required(e map[string]string, k, v string, min, max int) {
	n := len(strings.TrimSpace(v))
	if n < min {
		e[k] = "Polje je obavezno."
	} else if n > max {
		e[k] = "Vrednost je predugačka."
	}
}
