package models

import "testing"

func TestNormalizeImagesMigratesLegacyAndKeepsExactlyOnePrimary(t *testing.T) {
	images := NormalizeImages([]ProductImage{
		{URL: "https://example.com/second.jpg", LegacyStoragePath: "jasen-jela/products/p/second", Order: 4, IsPrimary: true},
		{URL: "https://example.com/first.jpg", PublicID: "jasen-jela/products/p/first", Order: 1, IsPrimary: true},
	})
	if images[0].PublicID != "jasen-jela/products/p/first" || images[0].Order != 0 || !images[0].IsPrimary {
		t.Fatalf("unexpected first image: %#v", images[0])
	}
	if images[1].PublicID != "jasen-jela/products/p/second" || images[1].LegacyStoragePath != "" || images[1].Order != 1 || images[1].IsPrimary {
		t.Fatalf("unexpected second image: %#v", images[1])
	}
}
