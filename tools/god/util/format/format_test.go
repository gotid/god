package format

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSplit(t *testing.T) {
	list, err := split("A")
	assert.Nil(t, err)
	assert.Equal(t, []string{"A"}, list)

	list, err = split("goDesigner")
	assert.Nil(t, err)
	assert.Equal(t, []string{"go", "Designer"}, list)

	list, err = split("God")
	assert.Nil(t, err)
	assert.Equal(t, []string{"God"}, list)

	list, err = split("go_designer")
	assert.Nil(t, err)
	assert.Equal(t, []string{"go", "designer"}, list)

	list, err = split("zhkGo_designer")
	assert.Nil(t, err)
	assert.Equal(t, []string{"zhk", "Go", "designer"}, list)

	list, err = split("GOD")
	assert.Nil(t, err)
	assert.Equal(t, []string{"G", "O", "D"}, list)

	list, err = split("god")
	assert.Nil(t, err)
	assert.Equal(t, []string{"god"}, list)

	list, err = split("")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(list))

	list, err = split("a_b_CD_EF")
	assert.Nil(t, err)
	assert.Equal(t, []string{"a", "b", "C", "D", "E", "F"}, list)

	list, err = split("_")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(list))

	list, err = split("__")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(list))

	list, err = split("_A")
	assert.Nil(t, err)
	assert.Equal(t, []string{"A"}, list)

	list, err = split("_A_")
	assert.Nil(t, err)
	assert.Equal(t, []string{"A"}, list)

	list, err = split("A_")
	assert.Nil(t, err)
	assert.Equal(t, []string{"A"}, list)

	list, err = split("welcome_to_go_designer")
	assert.Nil(t, err)
	assert.Equal(t, []string{"welcome", "to", "go", "designer"}, list)
}

func TestFileNamingFormat(t *testing.T) {
	testFileNamingFormat(t, "godesigner", "welcome_to_go_designer", "welcometogodesigner")
	testFileNamingFormat(t, "_go#designer_", "welcome_to_go_designer", "_welcome#to#go#designer_")
	testFileNamingFormat(t, "Go#designer", "welcome_to_go_designer", "Welcome#to#go#designer")
	testFileNamingFormat(t, "Go#Designer", "welcome_to_go_designer", "Welcome#To#Go#Designer")
	testFileNamingFormat(t, "Go_Designer", "welcome_to_go_designer", "Welcome_To_Go_Designer")
	testFileNamingFormat(t, "go_Designer", "welcome_to_go_designer", "welcome_To_Go_Designer")
	testFileNamingFormat(t, "goDesigner", "welcome_to_go_designer", "welcomeToGoDesigner")
	testFileNamingFormat(t, "GoDesigner", "welcome_to_go_designer", "WelcomeToGoDesigner")
	testFileNamingFormat(t, "GODesigner", "welcome_to_go_designer", "WELCOMEToGoDesigner")
	testFileNamingFormat(t, "GoDESIGNER", "welcome_to_go_designer", "WelcomeTOGODESIGNER")
	testFileNamingFormat(t, "GODESIGNER", "welcome_to_go_designer", "WELCOMETOGODESIGNER")
	testFileNamingFormat(t, "GO*DESIGNER", "welcome_to_go_designer", "WELCOME*TO*GO*DESIGNER")
	testFileNamingFormat(t, "[GO#DESIGNER]", "welcome_to_go_designer", "[WELCOME#TO#GO#DESIGNER]")
	testFileNamingFormat(t, "{go###designer}", "welcome_to_go_designer", "{welcome###to###go###designer}")
	testFileNamingFormat(t, "{go###designergo_designer}", "welcome_to_go_designer", "{welcome###to###go###designergo_designer}")
	testFileNamingFormat(t, "GogoDesignerdesigner", "welcome_to_go_designer", "WelcomegoTogoGogoDesignerdesigner")
	testFileNamingFormat(t, "前缀GoDesigner后缀", "welcome_to_go_designer", "前缀WelcomeToGoDesigner后缀")
	testFileNamingFormat(t, "GoDesigner", "welcometogodesigner", "Welcometogodesigner")
	testFileNamingFormat(t, "GoDesigner", "WelcomeToGoDesigner", "WelcomeToGoDesigner")
	testFileNamingFormat(t, "godesigner", "WelcomeToGoDesigner", "welcometogodesigner")
	testFileNamingFormat(t, "go_designer", "WelcomeToGoDesigner", "welcome_to_go_designer")
	testFileNamingFormat(t, "Go_Designer", "WelcomeToGoDesigner", "Welcome_To_Go_Designer")
	testFileNamingFormat(t, "Go_Designer", "", "")
	testFileNamingFormatErr(t, "go", "")
	testFileNamingFormatErr(t, "gODesigner", "")
	testFileNamingFormatErr(t, "designer", "")
	testFileNamingFormatErr(t, "goZEro", "welcome_to_go_designer")
	testFileNamingFormatErr(t, "goZERo", "welcome_to_go_designer")
	testFileNamingFormatErr(t, "designergo", "welcome_to_go_designer")
}

func testFileNamingFormat(t *testing.T, format, in, expected string) {
	format, err := FileNamingFormat(format, in)
	assert.Nil(t, err)
	assert.Equal(t, expected, format)
}

func testFileNamingFormatErr(t *testing.T, format, in string) {
	_, err := FileNamingFormat(format, in)
	assert.Error(t, err)
}
