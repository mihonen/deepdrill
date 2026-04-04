package engine


import (
	"strings"
	"testing"
    "github.com/PuerkitoBio/goquery"
)


func TestSemanticTree1(t *testing.T){

	html := `
	<body>
		<div> 
		  <div> 
		    <h1> 
		    Top Stories
		    </h1>
		    <div>
		      <div> 
		        <p> Trump this and that <p> 
		        <p> Another war </p>
		      </div>
		      <div>
		        <p> China  <p>
		        <p> Something crazy going on in china </p>
		      </div>
		    <div>
		  </div>
		</div>
	</body>
	`

	expected := `[group][0]
  [h1][00] Top Stories
  [group][01]
    [p][010] Trump this and that
    [p][011] Another war
  [group][02]
    [p][020] China
    [p][021] Something crazy going on in china`


	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
	    t.Errorf("%s", err)
	}

	semanticTree := CreateSemanticTree(doc)
	result := semanticTree.String()

	if result != expected {
	    t.Errorf("A: \n%s\n B: \n%s", expected, result)
	}
}




func TestSemanticTree2(t *testing.T){
	html := `
	<div>
	<div>
	<div>
	<div>
	<p> hola </p>
	<p> mista </p>
	</div>
	</div>
	</div>
	</div>
	`


	expected := `[group][0]
  [p][00] hola
  [p][01] mista`


	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
	    t.Errorf("%s", err)
	}

	semanticTree := CreateSemanticTree(doc)
	result := semanticTree.String()

	if result != expected {
	    t.Errorf("A: \n%q\n B: \n%q", expected, result)
	}

}




func TestSemanticTree3(t *testing.T){
	html := `
	<div>
	<nav><a href="/home"><div>Home</div> </a> <a href="/about"> <div>About</div> </a> </nav>
	<div>
	<img src="example.com/img/1.jpeg" alt="this is not a real image"/>
	<div>
	<div>
	<p> hola </p>
	<p> mista <span> magoo </span> </p>
	</div>
	<p></p>
	</div>
	</div>
	<p></p>
	</div>
	`


	expected := `[group][0]
  [link href=/home][00]
    [group][000] Home
  [link href=/about][01]
    [group][010] About
[group][1]
  [image alt=this is not a real image src=example.com/img/1.jpeg][10]
  [group][11]
    [p][110] hola
    [p][111] mista
      [span][1110] magoo`


	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
	    t.Errorf("%s", err)
	}

	semanticTree := CreateSemanticTree(doc)
	result := semanticTree.String()

	if result != expected {
	    t.Errorf("Expected: \n%s\n Got: \n%s", expected, result)
	}

}




func TestSemanticTree4(t *testing.T) {
	html := `
	<div>
		<header>
			<nav>
				<div class="wrapper">
					<a href="/"><div><span>Logo</span></div></a>
					<ul>
						<li><a href="/products">Products</a></li>
						<li><a href="/about">About</a></li>
					</ul>
				</div>
			</nav>
		</header>
		<main>
			<div class="hero">
				<div class="inner">
					<h1>Welcome to our site</h1>
					<p>We make great things</p>
				</div>
			</div>
			<div class="articles">
				<div class="article">
					<div class="meta">
						<time datetime="2024-01-01">January 1st</time>
						<a href="/author/john">John Doe</a>
					</div>
					<div class="content">
						<h2>First Article</h2>
						<p>Some intro text <span>with emphasis</span>here</p>
						<figure>
							<img src="article1.jpg" alt="Article image"/>
							<figcaption>This is the caption</figcaption>
						</figure>
					</div>
				</div>
				<div class="article">
					<div class="meta">
						<time datetime="2024-01-02">January 2nd</time>
					</div>
					<div class="content">
						<h2>Second Article</h2>
						<p>Another <span>interesting</span>piece</p>
						<video>
							<source src="video.mp4" type="video/mp4"/>
						</video>
					</div>
				</div>
			</div>
			<div class="empty">
				<div class="also-empty">
					<script>console.log("drop me")</script>
					<style>.drop { color: red }</style>
				</div>
			</div>
		</main>
	</div>
	`
	expected := `[group][0]
  [link href=/][00]
    [span][000] Logo
  [group][01]
    [link href=/products][010] Products
    [link href=/about][011] About
[group][1]
  [h1][10] Welcome to our site
  [p][11] We make great things
[group][2]
  [time datetime=2024-01-01][20] January 1st
  [link href=/author/john][21] John Doe
[group][3]
  [h2][30] First Article
  [p][31] Some intro text here
    [span][310] with emphasis
  [group][32]
    [image alt=Article image src=article1.jpg][320]
    [figcaption][321] This is the caption
[group][4]
  [time datetime=2024-01-02][40] January 2nd
  [group][41]
    [h2][410] Second Article
    [p][411] Another piece
      [span][4110] interesting
    [video poster= src=video.mp4 type=video/mp4][412]`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("failed to parse html: %v", err)
	}

	result := CreateSemanticTree(doc).String()
	if result != expected {
		t.Errorf("Expected:\n%s\n\n Got:\n%s", expected, result)
	}
}




