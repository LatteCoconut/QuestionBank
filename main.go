package main

import (
	"QuestionBank/SpiderUtil"
	"QuestionBank/xtheme"
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/signintech/gopdf"
	"strings"
)

//go:embed microsoftch.ttf
var fontBytes []byte

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("QuestionBank")
	myApp.Settings().SetTheme(&xtheme.XTheme{})

	ckEntry := widget.NewEntry()
	ckEntry.SetPlaceHolder("输入你的cookie信息....")
	ckEntry.MultiLine = true
	label := widget.NewLabel("Cookie")

	urlsEntry := widget.NewEntry()
	urlsEntry.SetPlaceHolder("输入你的错题页链接,多个则以英文逗号分隔...")
	urlsLabel := widget.NewLabel("Urls")

	// 创建一个无限进度条（加载动画）
	progressBar := widget.NewProgressBarInfinite()
	progressBar.Hide() // 初始时隐藏进度条

	grid := container.New(layout.NewFormLayout(), label, ckEntry, urlsLabel, urlsEntry)

	// 先声明一个*widget.Button类型的变量
	var submitBtn *widget.Button
	var pdfBtn *widget.Button
	questions := []SpiderUtil.Question{}

	submitBtn = widget.NewButton("submit", func() {
		cookieValue := ckEntry.Text
		urlsValue := urlsEntry.Text

		if cookieValue == "" || urlsValue == "" {
			dialog.ShowInformation("Warning", "Cookie value or URLs value cannot be empty!", myWindow)
			return
		}

		urlsList := strings.Split(urlsValue, ",")
		substr := "examRecord"
		for _, url := range urlsList {
			if !strings.Contains(url, substr) {
				dialog.ShowInformation("Warning", "输入的地址不对,请输入答题记录的链接...", myWindow)
				return
			}
		}

		cookieRaw, err := SpiderUtil.ParseCookie(cookieValue)
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}

		// 显示进度条并禁用按钮
		progressBar.Show()
		submitBtn.Disable()

		go func() {
			questions, err = SpiderUtil.ExecuteSpider(urlsList, cookieRaw)
			// 操作完成后，隐藏进度条并启用按钮
			progressBar.Hide()
			submitBtn.Enable()
			if err != nil {
				err := errors.New("Operation failed.")
				dialog.ShowError(err, myWindow)
				pdfBtn.Disable()
			} else {
				if len(questions) == 0 {
					err := errors.New("没有获取到数据，请重新登陆获取cookie.")
					dialog.ShowError(err, myWindow)
					pdfBtn.Disable()
					return
				}
				dialog.ShowInformation("Success", "Operation completed successfully.", myWindow)
				pdfBtn.Enable()
			}

		}()
	})

	pdfBtn = widget.NewButton("Export as PDF", func() {
		// 弹出文件保存对话框
		dialog.ShowFileSave(func(uc fyne.URIWriteCloser, err error) {
			if uc == nil { // 用户取消了操作
				return
			}
			if err != nil { // 处理可能的错误
				dialog.ShowError(err, myWindow)
				return
			}

			defer uc.Close()

			// 获取用户选择的文件路径
			filePath := uc.URI().Path()

			// 调用生成PDF的函数，并传入用户选择的文件路径
			err = exportQuestionsToPDF(questions, filePath)
			if err != nil {
				dialog.ShowError(err, myWindow)
			} else {
				dialog.ShowInformation("Success", "PDF has been saved successfully.", myWindow)
			}
		}, myWindow)
	})
	pdfBtn.Disable()
	buttonsContainer := container.NewHBox(layout.NewSpacer(), submitBtn, pdfBtn, layout.NewSpacer())

	// 使用容器来组合标签和输入框，以及下划线
	content := container.NewVBox(
		grid,
		buttonsContainer,
		progressBar,
	)

	myWindow.Resize(fyne.NewSize(900, 240))
	myWindow.SetContent(content)
	myWindow.ShowAndRun()

}

func exportQuestionsToPDF(questions []SpiderUtil.Question, filePath string) error {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.AddPage()

	// 加载嵌入的字体
	err := pdf.AddTTFFontByReader("MyFont", bytes.NewReader(fontBytes))
	if err != nil {
		return err
	}

	err = pdf.SetFont("MyFont", "", 10)
	if err != nil {
		return err
	}

	for i, q := range questions {
		currentY := pdf.GetY()
		if currentY > 780 { // 以800为界限，需要根据实际内容调整
			pdf.AddPage()
		}
		pdf.Br(10)
		// 计算MultiCell需要的Rect
		rect := &gopdf.Rect{W: 555, H: 30} // 宽度设置为页面宽度减去边距，高度暂时设为0

		title := fmt.Sprintf("%d、%s", i+1, q.Title)
		_ = pdf.MultiCell(rect, title)
		pdf.Br(5)

		optionsLine := strings.Join(q.Options, "  ")
		_ = pdf.MultiCell(rect, optionsLine)
		pdf.Br(5)

		correctText := fmt.Sprintf("答案: %s", q.Correct)
		_ = pdf.MultiCell(rect, correctText)
		pdf.Br(15)

	}

	// 输出PDF到文件
	err = pdf.WritePdf(filePath)
	if err != nil {
		return err
	}

	return nil
}
