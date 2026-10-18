package slides

// Settings are the reveal.js options a deck is presented with
type Settings struct {
	Theme                string         `json:"theme" bson:"theme" binding:"required,oneof=black white league beige night serif simple solarized moon dracula sky blood black-contrast white-contrast"`
	Transition           string         `json:"transition" bson:"transition" binding:"required,oneof=none fade slide convex concave zoom"`
	TransitionSpeed      string         `json:"transitionSpeed" bson:"transitionSpeed" binding:"required,oneof=default fast slow"`
	BackgroundTransition string         `json:"backgroundTransition" bson:"backgroundTransition" binding:"required,oneof=none fade slide convex concave zoom"`
	Controls             bool           `json:"controls" bson:"controls"`
	Progress             bool           `json:"progress" bson:"progress"`
	SlideNumber          bool           `json:"slideNumber" bson:"slideNumber"`
	Center               bool           `json:"center" bson:"center"`
	Loop                 bool           `json:"loop" bson:"loop"`
	AutoAnimate          bool           `json:"autoAnimate" bson:"autoAnimate"`
	AutoAnimateEasing    string         `json:"autoAnimateEasing" bson:"autoAnimateEasing" binding:"required,oneof=ease linear ease-in ease-out ease-in-out"`
	AutoAnimateDuration  float64        `json:"autoAnimateDuration" bson:"autoAnimateDuration" binding:"min=0.1,max=10"`
	Width                int            `json:"width" bson:"width" binding:"min=320,max=3840"`
	Height               int            `json:"height" bson:"height" binding:"min=240,max=2160"`
	Markdown             bool           `json:"markdown" bson:"markdown"`
	Highlight            bool           `json:"highlight" bson:"highlight"`
	HighlightTheme       string         `json:"highlightTheme" bson:"highlightTheme" binding:"required,oneof=monokai zenburn"`
	Editor               EditorSettings `json:"editor" bson:"editor"`
}

// EditorSettings set up the code editor of everyone working on the deck
type EditorSettings struct {
	Vim          bool   `json:"vim" bson:"vim"`
	LineWrapping bool   `json:"lineWrapping" bson:"lineWrapping"`
	LineNumbers  bool   `json:"lineNumbers" bson:"lineNumbers"`
	FontSize     int    `json:"fontSize" bson:"fontSize" binding:"min=10,max=24"`
	Indent       string `json:"indent" bson:"indent" binding:"required,oneof=tab 2 4"`
}

func DefaultEditorSettings() EditorSettings {
	return EditorSettings{LineWrapping: true, LineNumbers: true, FontSize: 14, Indent: "tab"}
}

func DefaultSettings() Settings {
	return Settings{
		Theme:                "black",
		Transition:           "slide",
		TransitionSpeed:      "default",
		BackgroundTransition: "fade",
		Controls:             true,
		Progress:             true,
		Center:               true,
		AutoAnimate:          true,
		AutoAnimateEasing:    "ease",
		AutoAnimateDuration:  1,
		Width:                960,
		Height:               700,
		Markdown:             true,
		Highlight:            true,
		HighlightTheme:       "monokai",
		Editor:               DefaultEditorSettings(),
	}
}
