package pdfparser
import ("bytes"; "fmt"; "math"; "strconv"; "strings"; "sort"; "encoding/hex"
	"github.com/signintech/gopdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)
//Комментарии все справа. Короткие строки соединяем через ; Секцию import тоже через ; где не github
//Никаких обилия пустых строк, Разделители перед функциями вида //______
const pointsToMm = 2.83465
type BBox struct { 	MinX, MinY, MaxX, MaxY float64; 	HasPoints bool;}
func (b *BBox) AddPoint(x, y float64) {
	if !b.HasPoints { b.MinX, b.MaxX, b.MinY, b.MaxY, b.HasPoints = x, x, y, y, true; return }
	if x < b.MinX { b.MinX = x }; 	if x > b.MaxX { b.MaxX = x }
	if y < b.MinY { b.MinY = y }; 	if y > b.MaxY { b.MaxY = y }
}
// RectResult описывает найденный элемент на странице
type RectResult struct { 	X, Y, Width, Height float64 }
type OverprintDetails struct { 
	HasStrokeOP, HasFillOP bool
	OPMValue               int }
//______________________________________________________________________________
// Generates PDF using gopdf with CMYK color, millimeter coordinates with bottom-left origin (0,0) and counter-clockwise rotation
func GenerateCMYKTextPDFBytes(fontPath string, text string, fontSize float64, x, y float64, angle float64, cmykHex string) ([]byte, error) { // Генерация PDF с координатами в ММ от ЛЕВОГО НИЖНЕГО угла и поворотом против часовой стрелки
	pdf := gopdf.GoPdf{} // Инициализация экземпляра PDF документа
	const mmToPt = 2.834645669291339 // Коэффициент перевода миллиметров в пункты (72/25.4)
	pageWidthPt := 1030.0 * mmToPt  // Ширина листа (1030 мм = 2919.685 pt)
	pageHeightPt := 770.0 * mmToPt  // Высота листа (770 мм = 2182.677 pt)
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: pageWidthPt, H: pageHeightPt}}) // Запуск документа с размером 1030x770 мм
	pdf.AddPage() // Добавление рабочей страницы
	err := pdf.AddTTFFont("custom_font", fontPath) // Загрузка шрифта TrueType
	if err != nil { return nil, fmt.Errorf("ошибка добавления шрифта: %v", err) } // Возврат ошибки при неудаче загрузки шрифта
	err = pdf.SetFont("custom_font", "", int(fontSize)) // Активация шрифта и размера
	if err != nil { return nil, fmt.Errorf("ошибка установки шрифта: %v", err) } // Возврат ошибки при установке шрифта
	if len(cmykHex) != 8 { return nil, fmt.Errorf("неверный формат CMYK hex, ожидается 8 символов (например, '000000FF')") } // Валидация строки цвета
	decodedBytes, err := hex.DecodeString(cmykHex) // Декодирование hex-строки
	if err != nil { return nil, fmt.Errorf("ошибка декодирования hex цвета: %v", err) } // Ошибка декодирования
	c, m, yVal, k := uint8(decodedBytes[0]), uint8(decodedBytes[1]), uint8(decodedBytes[2]), uint8(decodedBytes[3]) // Раскладка по каналам C, M, Y, K
	pdf.SetTextColorCMYK(c, m, yVal, k) // Установка CMYK цвета текста
	xPt := x * mmToPt // Перевод координаты X из миллиметров в пункты
	yPt := y * mmToPt // Перевод координаты Y из миллиметров в пункты
	gopdfYPt := pageHeightPt - yPt // Инверсия Y: отсчет от левого нижнего угла листа 770 мм
	if angle != 0 { pdf.Rotate(-angle, xPt, gopdfYPt) } // Поворот против часовой стрелки вокруг точки (x, y) в пунктах
	pdf.SetXY(xPt, gopdfYPt) // Установка позиции курсора в пунктах
	err = pdf.Text(text) // Вывод текста
	if err != nil {
		if angle != 0 { pdf.RotateReset() } // Сброс поворота при ошибке
		return nil, fmt.Errorf("ошибка вывода текста: %v", err) // Возврат ошибки вывода
	}
	if angle != 0 { pdf.RotateReset() } // Сброс поворота после отрисовки
	pdfBytes := pdf.GetBytesPdf() // Сборка документа в байтовый срез
	if len(pdfBytes) == 0 { return nil, fmt.Errorf("ошибка получения байтов PDF") } // Проверка на пустой результат
	return pdfBytes, nil // Возврат готовых байтов PDF
}//________________________________________________________________________________
// GetPdfBleedsForPage возвращает вылеты под обрез в мм: left, right, top, bottom, принимая PDF в виде байт
func GetPdfBleedsForPage(pdfData []byte, pageNum int) (left, right, top, bottom float64, err error) {
	rs := bytes.NewReader(pdfData)
	ctx, err := api.ReadContext(rs, model.NewDefaultConfiguration())
	if err != nil { return 0, 0, 0, 0, fmt.Errorf("ошибка открытия PDF из памяти: %w", err) }
	if err := ctx.EnsurePageCount(); err != nil { return 0, 0, 0, 0, fmt.Errorf("ошибка подсчета страниц: %w", err) }
	if pageNum < 1 || pageNum > ctx.PageCount { return 0, 0, 0, 0, fmt.Errorf("страница %d не существует (всего страниц: %d)", pageNum, ctx.PageCount) }
	// Запрашиваем границы конкретной страницы
	pageBoundaries, err := ctx.PageBoundaries(types.IntSet{pageNum: true})
	if err != nil { return 0, 0, 0, 0, fmt.Errorf("ошибка при получении границ: %w", err) }
	// Извлекаем границы
	var pageBounds model.PageBoundaries
	for _, bounds := range pageBoundaries { pageBounds = bounds; break }
	tb, bb := pageBounds.TrimBox(), pageBounds.BleedBox()
	// Если BleedBox не задан явно — берем MediaBox в качестве внешнего контейнера
	if bb == nil { bb = pageBounds.MediaBox() }
	if tb == nil || bb == nil { return 0, 0, 0, 0, fmt.Errorf("в PDF отсутствуют необходимые метки (TrimBox или Bleed/MediaBox)") }
	// Расчет вылетов для каждой стороны в мм
	left = (tb.LL.X - bb.LL.X) / pointsToMm; 	right = (bb.UR.X - tb.UR.X) / pointsToMm
	top = (bb.UR.Y - tb.UR.Y) / pointsToMm;  	bottom = (tb.LL.Y - bb.LL.Y) / pointsToMm
	return left, right, top, bottom, nil
}
//________________________________________________________________________________
// FindLowestLeftQurve3 читает PDF из памяти и  находит самый левый нижний объект с указанной заливкой и обводкой, возвращает координаты левого нижнего угла и расстояние между ближайшими обьектами по x,y.
func FindLowestLeftQurve3(pdfData []byte, targetPage int, fillFilter, strokeFilter string) (x1, y1, widthCalc, heightCalc float64, numX, numY int, gap float64, found bool, err error) {
	results, err := FindQurveByColors(pdfData, targetPage, fillFilter, strokeFilter)
	if err != nil || len(results) == 0 { return 0, 0, 0, 0, 0, 0, 0, false, err }	
	lowestLeftItem := results[0] // Находим самый левый нижний объект
	for _, item := range results { if item.X < lowestLeftItem.X || (item.X == lowestLeftItem.X && item.Y < lowestLeftItem.Y) { lowestLeftItem = item }	
	}
	x1, y1 = lowestLeftItem.X, lowestLeftItem.Y
	hasRight, hasAbove := false, false
	minRightX, minAboveY := math.MaxFloat64, math.MaxFloat64
	eps := 0.001
	for _, item := range results {		
		if math.Abs(item.X-x1) < eps && math.Abs(item.Y-y1) < eps { continue } // Исключаем сам базовый объект		
		if item.X > x1 && item.X < minRightX { minRightX, hasRight = item.X, true } // Ищем ближайший объект правее		
		if item.Y > y1 && item.Y < minAboveY { minAboveY, hasAbove = item.Y, true } // Ищем ближайший объект выше
	}	
	if hasRight { widthCalc = minRightX - x1 } else { widthCalc = lowestLeftItem.Width } // Если справа ничего нет — берем ширину нашего объекта	
	if hasAbove { heightCalc = minAboveY - y1 } else { heightCalc = lowestLeftItem.Height } // Если выше ничего нет — берем высоту нашего объекта
	targetRoundedX, targetRoundedY := math.Round(x1), math.Round(y1)
	var xValues []float64
	for _, item := range results {
		if math.Round(item.Y) == targetRoundedY { numX++ }
		if math.Round(item.X) == targetRoundedX { numY++ }
		if math.Round(item.Y) == targetRoundedY { xValues = append(xValues, item.X) }
	}
	sort.Float64s(xValues)
	var widthX_1_2, widthXcenter float64
	if len(xValues) >= 2 { widthX_1_2 = xValues[1] - xValues[0] }
	numXcenter := numX / 2 - 1
	if numXcenter >= 0 && numXcenter < len(xValues) && (numXcenter+1) < len(xValues) {
		widthXcenter = xValues[numXcenter+1] - xValues[numXcenter]
	}
	gap = widthXcenter - widthX_1_2  	//fmt.Printf("_parser____\n  %.2f\n  %.2f\n  %.2f\n  %.2f\n  %d\n  %d\n  %.2f\n  %.2f\n =====\n", x1, y1, widthCalc, heightCalc, numX, numY, xValues[numXcenter], xValues[numXcenter+1])
	return x1, y1, widthCalc, heightCalc, numX, numY, gap, true, nil
}
//________________________________________________________________________________
// FindLowestLeftQurve читает PDF из памяти и  находит самый левый нижний объект с указанной заливкой и обводкой, возвращает координаты левого нижнего угла и ширину высоту обьекта.
func FindLowestLeftQurve(pdfData []byte, targetPage int, fillFilter, strokeFilter string) (x1, y1, wMM, hMM float64, found bool, err error) {
	results, err := FindQurveByColors(pdfData, targetPage, fillFilter, strokeFilter)
	if err != nil || len(results) == 0 { return 0, 0, 0, 0, false, err }
	lowestLeftItem := results[0]
	for _, item := range results { if item.X < lowestLeftItem.X || (item.X == lowestLeftItem.X && item.Y < lowestLeftItem.Y) { lowestLeftItem = item } 	}
	return lowestLeftItem.X, lowestLeftItem.Y, lowestLeftItem.Width, lowestLeftItem.Height, true, nil
}
//________________________________________________________________________________
// FindQurveByColors читает PDF из памяти и находит все кривые с указанной заливкой и обводкой, а возвращает координаты  левого нижнего угла и ширину и высоту
func FindQurveByColors(pdfData []byte, targetPage int, fillFilter, strokeFilter string) ([]RectResult, error) {
	rs := bytes.NewReader(pdfData)
	ctx, err := api.ReadContext(rs, model.NewDefaultConfiguration())
	if err != nil { return nil, fmt.Errorf("ошибка чтения PDF из памяти: %v", err) }
	// Инициализируем страницы, чтобы pdfcpu корректно построил внутреннее дерево
	if err := ctx.EnsurePageCount(); err != nil { return nil, fmt.Errorf("ошибка инициализации страниц: %v", err) }
	pageDict, _, _, err := ctx.PageDict(targetPage, false)
	if err != nil { return nil, fmt.Errorf("ошибка получения словаря страницы: %v", err) }
	contentStream, err := ctx.PageContent(pageDict, targetPage)
	if err != nil { return nil, fmt.Errorf("ошибка контента страницы: %v", err) }
	overprintStates := make(map[string]OverprintDetails)
	if resDict := pageDict.DictEntry("Resources"); resDict != nil {
		if gsDict := resDict.DictEntry("ExtGState"); gsDict != nil {
			for gsName, gsObj := range gsDict {
				var subDict types.Dict
				if d, ok := gsObj.(types.Dict); ok { subDict = d } else if indRef, ok := gsObj.(types.IndirectRef); ok {
					if derefObj, err := ctx.Dereference(indRef); err == nil {
						if d, ok := derefObj.(types.Dict); ok { subDict = d }
					}
				}

				if subDict != nil { details := OverprintDetails{}
					if opObj, ok := subDict["OP"]; ok { if b, isBool := opObj.(types.Boolean); isBool && bool(b) { details.HasStrokeOP = true } }
					if opObj, ok := subDict["op"]; ok { if b, isBool := opObj.(types.Boolean); isBool && bool(b) { details.HasFillOP = true } }
					if opmObj, ok := subDict["OPM"]; ok { if i, isInt := opmObj.(types.Integer); isInt { details.OPMValue = int(i) } }
					overprintStates["/"+gsName] = details
				}
			}
		}
	}

	tokens := strings.Fields(string(contentStream))
	var stack []string
	var currentBBox BBox
	var results []RectResult

	var currStrokeC, currStrokeM, currStrokeY, currStrokeK float64
	var currFillC, currFillM, currFillY, currFillK float64
	var gsStrokeOP, gsFillOP, opOpStroke, opOpFill bool
	var gsOPM int

	const ptToMM = 0.352778

	for _, token := range tokens {
		switch token {	case "gs":
			if len(stack) >= 1 {
				if details, ok := overprintStates[stack[len(stack)-1]]; ok { gsStrokeOP, gsFillOP, gsOPM = details.HasStrokeOP, details.HasFillOP, details.OPMValue }
			}
			stack = nil

		case "OP": if len(stack) >= 1 { opOpStroke = (stack[len(stack)-1] == "true") }; stack = nil
		case "op": if len(stack) >= 1 { opOpFill = (stack[len(stack)-1] == "true") }; stack = nil
		case "K":  if len(stack) >= 4 { currStrokeC, currStrokeM, currStrokeY, currStrokeK = parseNum(stack[len(stack)-4]), parseNum(stack[len(stack)-3]), parseNum(stack[len(stack)-2]), parseNum(stack[len(stack)-1]) }; 	stack = nil
		case "SC", "SCN":
			if len(stack) >= 4 {
				currStrokeC, currStrokeM, currStrokeY, currStrokeK = parseNum(stack[len(stack)-4]), parseNum(stack[len(stack)-3]), parseNum(stack[len(stack)-2]), parseNum(stack[len(stack)-1])
			} else if len(stack) >= 1 && parseNum(stack[len(stack)-1]) == 0 {
				currStrokeC, currStrokeM, currStrokeY, currStrokeK = 0, 0, 0, 0 }; 			stack = nil

		case "k": if len(stack) >= 4 { currFillC, currFillM, currFillY, currFillK = parseNum(stack[len(stack)-4]), parseNum(stack[len(stack)-3]), parseNum(stack[len(stack)-2]), parseNum(stack[len(stack)-1]) }
			stack = nil

		case "sc", "scn":
			if len(stack) >= 4 {
				currFillC, currFillM, currFillY, currFillK = parseNum(stack[len(stack)-4]), parseNum(stack[len(stack)-3]), parseNum(stack[len(stack)-2]), parseNum(stack[len(stack)-1])
			} else if len(stack) >= 1 && parseNum(stack[len(stack)-1]) == 0 {
				currFillC, currFillM, currFillY, currFillK = 0, 0, 0, 0 	}
			stack = nil

		case "G": if len(stack) >= 1 { currStrokeC, currStrokeM, currStrokeY, currStrokeK = 0, 0, 0, 1.0-parseNum(stack[len(stack)-1]) }; stack = nil
		case "g":	if len(stack) >= 1 { currFillC, currFillM, currFillY, currFillK = 0, 0, 0, 1.0-parseNum(stack[len(stack)-1]) }; stack = nil
		case "re": 
			if len(stack) >= 4 {
				x, y, w, h := parseNum(stack[len(stack)-4]), parseNum(stack[len(stack)-3]), parseNum(stack[len(stack)-2]), parseNum(stack[len(stack)-1])
				currentBBox = BBox{}
				currentBBox.AddPoint(x, y)
				currentBBox.AddPoint(x+w, y+h) 	}; 			stack = nil

		case "m", "l", "v", "y": 	if len(stack) >= 2 { currentBBox.AddPoint(parseNum(stack[len(stack)-2]), parseNum(stack[len(stack)-1])) }; 		stack = nil
		case "c": if len(stack) >= 6 { currentBBox.AddPoint(parseNum(stack[len(stack)-2]), parseNum(stack[len(stack)-1])) };		stack = nil
		case "S", "s", "f", "F", "B", "B*", "b", "b*":
			if currentBBox.HasPoints {
				isStroke := (token == "S" || token == "s" || token == "B" || token == "B*" || token == "b" || token == "b*")
				isFill := (token == "f" || token == "F" || token == "B" || token == "B*" || token == "b" || token == "b*")
				fillMatched := matchColor(fillFilter, isFill, currFillC, currFillM, currFillY, currFillK, opOpFill || gsFillOP)
				strokeMatched := matchColor(strokeFilter, isStroke, currStrokeC, currStrokeM, currStrokeY, currStrokeK, opOpStroke || gsStrokeOP || gsOPM > 0)
				if fillMatched && strokeMatched {
					results = append(results, RectResult{
						X: currentBBox.MinX * ptToMM, Y: currentBBox.MinY * ptToMM,
						Width: (currentBBox.MaxX - currentBBox.MinX) * ptToMM, Height: (currentBBox.MaxY - currentBBox.MinY) * ptToMM,
					})
				}
			}
			currentBBox, stack = BBox{}, nil

		case "n":	currentBBox, stack = BBox{}, nil
		default: 	stack = append(stack, token)
		}
	}

	return results, nil
}
//________________________________________________________________________________
func matchColor(filter string, isActive bool, c, m, y, k float64, opActive bool) bool {
	if filter == "*" { return true }
	if filter == "" { return !isActive }
	if !isActive { return false }
	parts := strings.Split(filter, "-")
	if len(parts) < 4 { return false }

	fc, fm, fy, fk := parseNum(parts[0]), parseNum(parts[1]), parseNum(parts[2]), parseNum(parts[3])
	if fc > 1 || fm > 1 || fy > 1 || fk > 1 { fc, fm, fy, fk = fc/100, fm/100, fy/100, fk/100 }
	eps := 0.01
	if math.Abs(c-fc) > eps || math.Abs(m-fm) > eps || math.Abs(y-fy) > eps || math.Abs(k-fk) > eps { return false }
	return (len(parts) == 5 && parts[4] == "o") == opActive
}

func parseNum(s string) float64 { val, _ := strconv.ParseFloat(s, 64);  return val }