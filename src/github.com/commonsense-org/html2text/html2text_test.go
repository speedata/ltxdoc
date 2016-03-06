package html2text

import (
	"testing"
)

func TestText(t *testing.T) {
	testCases := []struct {
		input string
		expr  string
	}{
		{
			`<li>
  <a href="/new" data-ga-click="Header, create new repository, icon:repo"><span class="octicon octicon-repo"></span> New repository</a>
</li>`,
			"* [New repository](http://www.microshwhat.com/bar/soapy/new)",
		},
		{
			`hi1

			<br>

	hello <a href="https://google.com">google</a>
	<br><br>
	test<p>List:</p>

	<ul>
		<li><a href="foo">Foo</a></li>
		<li><a href="http://www.microshwhat.com/bar/soapy">Barsoap</a></li>
        <li>Baz</li>
	</ul>
`,
			"hi1 \n\n hello [google](https://google.com)\n\n test\n\nList:\n\n    * [Foo](foo)\n    * [Barsoap](http://www.microshwhat.com/bar/soapy)\n    * Baz",
		},
		// Malformed input html.
		{
			`hi2

			hello <a href="https://google.com">google</a>

			test<p>List:</p>

			<ul>
				<li><a href="foo">Foo</a>
				<li><a href="/
		                bar/baz">Bar</a>
		        <li>Baz</li>
			</ul>
		`,
			"hi2 hello [google](https://google.com) test\n\nList:\n\n    * [Foo](foo)\n    * [Bar](http://www.microshwhat.com/bar/soapy/bar/baz)\n    * Baz",
		},
		{
			`<table><tr><th>First Column</th><th>Second Column</th><th>Third Column</th></tr>
				<tr><td>row1column1</td><td>row1column2</td><td>row1column3</td></tr>
				<tr><td>row2column1</td><td>row2column2</td><td>row2column3</td></tr>
				<tr><td>row3column1</td><td>row3column2</td><td>row3column3</td></tr>
				</table>`,
			"First Column: row1column1\nSecond Column: row1column2\nThird Column: row1column3\n\nFirst Column: row2column1\nSecond Column: row2column2\nThird Column: row2column3\n\nFirst Column: row3column1\nSecond Column: row3column2\nThird Column: row3column3",
		},
		{`<script>$(document).ready(function(){alert("this");});</script>`,
			"",
		},
	}

	for i, testCase := range testCases {
		text, err := FromString(testCase.input, "http://www.microshwhat.com/bar/soapy", OmitClasses|OmitIds|OmitRoles)
		if err != nil {
			t.Error(err)
		}

		if testCase.expr != text {
			t.Errorf("FAILED[%d]:\nOutput:\n>>>>\n%s\n<<<<\nExpression:\n>>>>\n%s\n<<<<\n", i, text, testCase.expr)
		} else {
			t.Logf("PASSED[%d]:\ninput:\n\n%s\n\n\n\noutput:\n\n%s\n", i, testCase.input, text)
		}
	}
}
