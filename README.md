<h1 align="center"> CAST-text </h1>
<p align="center"> A zero latency, easy-to-use full-text rss terminal reader.</p>

<div align="center"> <img width="640" src="https://github.com/user-attachments/assets/25d765b8-51fc-4b58-8d10-db82d93265c1"/> </div>

## Features
 - Read the full article content from the terminal
 - Easy to use: ←↓↑→ arrow keys (or hjkl) is all you need to use its main features. Up/Dow to browse through articles, Left/Right to scroll the article content. 
 - Fast: it will prefetch adjacent articles - so every action is instant.
 - By default it will load the BBC, but you can change the source by passing `-rss <your_rss_or_atom_feed>`
 - One single UI for both article listing and article content.
 - If for some reason you need to open the article on a browser simply press ENTER.
 - q to quit, other optinal shortcuts put in ( ) on the title to jump directly to the article
 - Trackpad scrolling will list through articles (thanks to the great rivo/tview library)

## Showcase
https://github.com/user-attachments/assets/90f3fe3e-a555-42d2-b071-7cbab2cb3172
### Default frontpage is BBC
<img width="1412" alt="image" src="https://github.com/user-attachments/assets/a83f210b-a359-4541-bb95-83dbe8ec3094">

# Installation
## Homebrew
```
brew tap piqoni/cast-text
brew install cast-text
```

## Binaries
Download binaries for your OS at [release page](https://github.com/piqoni/cast-text/releases), and chmod +x the file to allow execution. 

## Using GO INSTALL
If you use GO, you can install it directly:
```
go install github.com/piqoni/cast-text@latest
```

# How to run it
If you want to read BBC just run the binary. If you want to read something else pass its rss address to -rss parameter. For example to read http://lobste.rs start it by running:
```
cast-text -rss https://lobste.rs/rss
```
You can perhaps put an alias (export alias lobster=cast-text -rss https://lobste.rs/rss) so now you have a "lobster" reader shorthand:

<img width="895" alt="image" src="https://github.com/user-attachments/assets/a6b26ceb-50b2-4ff6-9e7c-c43aaa142208">

