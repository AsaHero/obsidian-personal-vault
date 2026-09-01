# How I Built a Multilingual City Search Engine for the World

## Task:

**Enable users to search for any city in the world using city, state, or country names — in any language.**

Sounds simple, right?

At first glance, it’s just a matter of collecting some city names and running a search. But as I dove into the problem, I found myself in a rabbit hole of linguistic nuance, data messiness, and performance pitfalls. What started as a weekend project turned into a fascinating journey across transliteration, language detection, semantic embeddings, and full-text search engines.

So here’s the full story — a breakdown of the major issues I faced, and how I overcame them to build a truly **multilingual**, **semantic**, and **performant** global city search engine.

---

## 1. Step One: Gather Some Cities

The foundation, of course, is the data.

I grabbed a city dataset from Kaggle, fired up a Jupyter Notebook, and began the classic drill:

- Clean up missing values
    
- Standardize names
    
- Attach latitudes, longitudes, countries, and state codes
    
- Export to PostgreSQL
    

In the end, I had a polished dataset of **47,000+ cities**, each with their metadata. Great start.

But — all names were in Latin script, mostly using "universal" spellings (e.g., "Moscow", not "Москва").

So, the best I could do at this point was implement a simple `ILIKE` search or PostgreSQL's `ts_query`. I built a basic Golang server, added a search endpoint, threw together some HTML/JS frontend, and boom — search works. But **not multilingual yet**.

---

## 2. Let’s Add Transliteration

Next idea: **transliterate** user queries to match my Latin-script database.

For example:

- User types “Лондон” (Russian for London)
    
- We transliterate → “London”
    
- And voilà, a match!
    

The process is called **transliteration**, and it maps characters from one script to another without translating the meaning. I explored libraries in multiple languages — Go, Python, C++, Java — and quickly found that **Python has the best tools** for this.

I ended up using **[PyICU](https://pypi.org/project/PyICU/)** — a powerful library with the `"Any-Latin"` mode that works impressively well for converting scripts. However, transliteration quality improves significantly if you know the language of the text.

So, I added **language detection** to the pipeline using **gcld3** (Google’s Compact Language Detector v3). It's fast, local (no external API), and decently accurate. I opted for the `gcdl3` Python wrapper — pure Python, easy to use, no C-extension pain.

Then I built a simple **FastAPI + Docker** microservice for transliteration + language detection, and integrated it into my Go backend.

So far so good — users could now search for “Токио”, “Londres”, or “برلين”, and my system would try to transliterate and find the correct Latin form.

---

## 3. Problem: “Китай” is not “Kitai”

Here’s where things got messy.

Transliteration only works when the **name is structurally similar across languages**.

For example:

- "Москва" → "Moskva" — fine
    
- "Китай" → "Kitai" — not fine
    

Because "Kitai" isn’t a city in my database — the user actually meant **"China"**, a **country**, and the names don’t even resemble each other in different languages.

That’s when I realized:  
**Transliteration alone isn't enough.**  
I needed **alternate names** in different languages.

---

## 4. Alternate Names to the Rescue

After some digging, I found **GeoNames** — a massive geolocation database with alternate names in dozens of languages. It was a mess of a dataset, but after a long night with `pandas` and `seaborn`, I filtered out the noise and mapped alternate names to my cities.

Now, “Beijing” was also searchable as:

- 北京 (Chinese)
    
- Pekín (Spanish)
    
- Pékin (French)
    
- Пекин (Russian)
    

This was a **huge step forward** in making my search multilingual.

I updated my database schema to store alternate names, indexed them, and updated my search logic to query both the main name and all known alternates.

But… it got **very slow**.

---

## 5. Enter: Typesense

To fix the performance issues, I turned to **Typesense** — an open-source, in-memory, typo-tolerant search engine that’s perfect for fast keyword lookups.

Within an hour, I had it running with Docker. I:

- Defined a city schema
    
- Migrated all city data + alternate names
    
- Integrated it into my backend
    

**Boom. Instant performance boost.**  
Now I had:

- Fast full-text search
    
- Built-in fuzzy matching (for typos)
    
- Real-time response times
    

This solved most of the performance and misspelling issues. But — the “China” problem still persisted, especially for country and state names.

---

## 6. The Final Boss: Semantics

Even with transliteration and alternate names, users searching “Китай” would get nothing — because it's a **semantic mismatch**, not a spelling issue.

So I took it a step further:

### I added **semantic search** using **vector embeddings**.

---

## 7. Hybrid Vector + Keyword Search

Here's what I did:

1. **Generate Embeddings:**  
    I used a multilingual sentence transformer (like `distiluse-base-multilingual-cased-v1`) to embed all:
    
    - City names
        
    - State names
        
    - Country names
        
    - Alternate names
        
    - Transliterated forms
        
2. **Update Typesense Schema:**  
    Each document in Typesense now included an `embedding` vector alongside its name fields.
    
3. **Hybrid Search Logic:**  
    I enabled Typesense's hybrid search mode — which allows you to combine:
    
    - Keyword-based ranking (fuzzy match, typo tolerance)
        
    - Vector similarity ranking (semantic closeness)
        
4. **Custom Score Fusion:**  
    I added a custom fusion formula that weights both:
    
    - `text_match_score` (from keyword)
        
    - `vector_distance` (from semantic similarity)
        
    
    So even if a query doesn’t match exactly, it can still surface relevant results based on meaning.
    

---

## 8. Final Result

- User types: `Китай`
    
- System transliterates: `Kitai`
    
- Keyword match fails
    
- Vector match kicks in: finds **China**
    
- Boom: **"China"** and its cities now show up
    

Thanks to a combination of:

- Transliteration
    
- Language detection
    
- Alternate names
    
- Semantic vector search
    
- Typesense hybrid scoring
    

…I now have a **robust**, **fast**, and **truly multilingual** city search engine that works **across scripts, languages, and cultural nuances.**

---

## What’s Next?

I’m considering:

- Expanding to place types (mountains, rivers, landmarks)
    
- Supporting dialects and regional variations
    
- Weighting results based on population or popularity
    

But for now, I’m proud of what this turned into — a city search that finally feels global.

---

Let me know if you want this in Markdown for your blog, LinkedIn post, or as a script for a YouTube breakdown — I can prep it!