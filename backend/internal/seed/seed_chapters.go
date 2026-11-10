package seed

import (
	"github.com/guitarkeegan/music-class-tutor/internal/data"
)

// seedBook is a public domain book from Project Gutenberg whose chapter text
// lives in books/<dir>/<dir>-<author>-<edition>-<chapter>.txt
type seedBook struct {
	book     data.Book
	chapters []data.Chapter
}

func seedBooks() []seedBook {
	return []seedBook{
		{
			book: data.Book{Title: "Alice in Wonderland", Edition: 3, Year: 1865},
			chapters: []data.Chapter{
				{ChapterNumber: 1, Title: "Down the Rabbit-Hole"},
				{ChapterNumber: 2, Title: "The Pool of Tears"},
				{ChapterNumber: 3, Title: "A Caucus-Race and a Long Tale"},
				{ChapterNumber: 4, Title: "The Rabbit Sends in a Little Bill"},
				{ChapterNumber: 5, Title: "Advice from a Caterpillar"},
				{ChapterNumber: 6, Title: "Pig and Pepper"},
				{ChapterNumber: 7, Title: "A Mad Tea-Party"},
				{ChapterNumber: 8, Title: "The Queen’s Croquet-Ground"},
				{ChapterNumber: 9, Title: "The Mock Turtle’s Story"},
				{ChapterNumber: 10, Title: "The Lobster Quadrille"},
				{ChapterNumber: 11, Title: "Who Stole the Tarts?"},
				{ChapterNumber: 12, Title: "Alice’s Evidence"},
			},
		},
		{
			book: data.Book{Title: "Life of Chopin", Edition: 4, Year: 1880},
			chapters: []data.Chapter{
				{ChapterNumber: 1, Title: "Chopin—Style and Improvements—The Adagio of the Second Concerto—Funeral March—Psychological Character of the Compositions of Chopin, &c., &c."},
				{ChapterNumber: 2, Title: "National Character of the Polonaise—Oginski—Meyseder—Weber—Chopin—His Polonaise in F Sharp Minor—Polonaise—Fantaisie"},
				{ChapterNumber: 3, Title: "Chopin's Mazourkas—Polish Ladies—Mazourka in Poland—Tortured Motives—Early Life of Chopin—Zal"},
				{ChapterNumber: 4, Title: "Chopin's Mode of Playing—Concerts—The Elite—Fading Bouquets and Immortal Crowns—Hospitality—Heine—Meyerbeer—Adolphe Nourrit—Eugene Delacroix—Niemcevicz—Mickiewicz—George Sand"},
				{ChapterNumber: 5, Title: "The Lives of Artists—Pure Fame of Chopin—Reserve—Classic and Romantic Art—Language of the Sclaves—Chopin's Love of Home Memories"},
				{ChapterNumber: 6, Title: "Birth and Early Life of Chopin—National Artists—Chopin Embodies in Himself the Poetic Sense of His Whole Nation—Opinion of Beethoven"},
				{ChapterNumber: 7, Title: "Madame Sand—Lelia—Visit to Majorca—Exclusive Ideals"},
				{ChapterNumber: 8, Title: "Disappointment—Ill Health—Visit to England—Devotion of Friends—Last Sacraments—Delphina Potocka—Louise—M. Gutman—Death"},
			},
		},
	}
}
