import { ArrowRight, ArrowUpRight } from "lucide-react";

// Editorial starting points. Selecting one searches the live catalog.
const picks = [
  {
    title: "The Creative Act",
    author: "Rick Rubin",
    isbn: "9780593652886",
    note: "Make room for a different way of seeing.",
    color: "#c7c1b3",
  },
  {
    title: "The Midnight Library",
    author: "Matt Haig",
    isbn: "9780525559474",
    note: "A little possibility. A whole other life.",
    color: "#b9c4d0",
  },
  {
    title: "Dune",
    author: "Frank Herbert",
    isbn: "9780441172719",
    note: "A world worth getting lost in.",
    color: "#d3c5b4",
  },
  {
    title: "A Room of One’s Own",
    author: "Virginia Woolf",
    isbn: "9780156787338",
    note: "An enduring argument for creative freedom.",
    color: "#c9bfce",
  },
];
export function Discovery({
  onSearch,
  onCollection,
}: {
  onSearch: (query: string) => void;
  onCollection: () => void;
}) {
  return (
    <div className="discovery-content">
      <section className="editorial-section" aria-labelledby="reading-picks">
        <div className="section-heading">
          <div>
            <h2 id="reading-picks">A few good places to begin</h2>
            <p>Selected for your next reading mood.</p>
          </div>
          <span className="editorial-caption">The Imprint selection</span>
        </div>
        <div className="editorial-grid">
          {picks.map((book) => (
            <button
              className="editorial-book"
              key={book.isbn}
              onClick={() => onSearch(book.title)}
            >
              <span
                className="editorial-cover"
                style={{ backgroundColor: book.color }}
              >
                <img
                  src={`https://covers.openlibrary.org/b/isbn/${book.isbn}-L.jpg?default=false`}
                  alt={`${book.title} cover`}
                  loading="lazy"
                  onError={(e) => {
                    e.currentTarget.style.visibility = "hidden";
                  }}
                />
                <span className="editorial-cover-title">{book.title}</span>
                <ArrowUpRight size={18} className="editorial-arrow" />
              </span>
              <span className="editorial-title">{book.title}</span>
              <span className="editorial-author">{book.author}</span>
              <span className="editorial-note">{book.note}</span>
            </button>
          ))}
        </div>
      </section>
      <section className="browse-section" aria-labelledby="browse-mood">
        <div>
          <h2 id="browse-mood">Follow your curiosity</h2>
          <p>A familiar favorite or somewhere entirely new.</p>
        </div>
        <div className="genre-links">
          {[
            "Literary fiction",
            "Science fiction",
            "Mystery",
            "Fantasy",
            "History",
            "Philosophy",
            "Poetry",
            "Biography",
          ].map((genre) => (
            <button key={genre} onClick={() => onSearch(genre)}>
              {genre}
              <ArrowUpRight size={15} />
            </button>
          ))}
        </div>
      </section>
      <section className="library-invitation">
        <div>
          <h2>A home for your reading life.</h2>
          <p>Keep your next reads, favorite editions, and thoughts together.</p>
        </div>
        <button className="button primary" onClick={onCollection}>
          Explore your collection <ArrowRight size={17} />
        </button>
      </section>
    </div>
  );
}
