// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.707
package web

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

func Login() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, templ_7745c5c3_W io.Writer) (templ_7745c5c3_Err error) {
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templ_7745c5c3_W.(*bytes.Buffer)
		if !templ_7745c5c3_IsBuffer {
			templ_7745c5c3_Buffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templ_7745c5c3_Buffer)
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"login-page\"><h2>Login</h2><form hx-post=\"/login\" hx-target=\"#app\" hx-swap=\"outerHTML\"><div class=\"form-group\"><label for=\"email\">Email</label> <input type=\"email\" class=\"form-control\" id=\"email\" name=\"email\" placeholder=\"Email\" required></div><div class=\"form-group\"><label for=\"password\">Password</label> <input type=\"password\" class=\"form-control\" id=\"password\" name=\"password\" placeholder=\"Password\" required></div><button type=\"submit\" class=\"btn btn-primary\">Login</button></form><p class=\"mt-3\">Don't have an account? <a href=\"/register\" hx-get=\"/register\" hx-target=\"#app\" hx-swap=\"outerHTML\">Register here</a></p></div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if !templ_7745c5c3_IsBuffer {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteTo(templ_7745c5c3_W)
		}
		return templ_7745c5c3_Err
	})
}
