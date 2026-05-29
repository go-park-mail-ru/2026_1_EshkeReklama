package service

import (
	"fmt"
	"strings"
)

type AIPromptMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AIPromptSet struct {
	Messages []AIPromptMessage `json:"messages"`
}

func BuildGenerateAdTextPrompt(in GenerateAdTextInput) AIPromptSet {
	in = normalizeGenerateAdTextInput(in)
	productName := strings.TrimSpace(in.ProductName)
	if productName == "" {
		productName = "не указано"
	}

	system := `Ты опытный рекламный копирайтер. Твоя задача — создавать короткие, понятные и убедительные рекламные объявления на русском языке для товаров, услуг, заведений и digital-продуктов.

Правила:
- Пиши только на русском языке.
- Сохраняй рекламный стиль, но без кликбейта и агрессивных манипуляций.
- Не выдумывай факты, скидки, цифры, гарантии или свойства, которых нет во входных данных.
- Не используй запрещённые или сомнительные формулировки: "лучший", "гарантированно", "100%", "навсегда", если этого нет во входе.
- Заголовок должен быть коротким, цепким и ясным.
- Описание должно объяснять пользу продукта простым языком.
- Если рекламируется физическое место, заведение или офлайн-услуга, делай акцент на атмосфере, удобстве, впечатлении и реальной ценности для посетителя.
- Если рекламируется товар, делай акцент на пользе, особенностях применения и понятном результате для клиента.
- Если рекламируется digital-продукт или сервис, делай акцент на выгоде, удобстве, экономии времени, росте эффективности или понятном результате.
- Не добавляй лишние пояснения, комментарии, markdown или кавычки вокруг всего ответа.
- Верни результат строго в JSON-формате с полями:
  {
    "headline": "строка",
    "body": "строка"
  }
- Длина headline не должна превышать ограничение пользователя.
- Длина body не должна превышать ограничение пользователя.
- Если входных данных недостаточно, всё равно сформируй максимально нейтральный и безопасный рекламный вариант без выдуманных деталей.`

	user := fmt.Sprintf(`Сгенерируй рекламное объявление.

Данные о продукте:
Название продукта: %s
Описание товара или услуги: %s
Тон: %s

Ограничения:
Максимальная длина заголовка: %d
Максимальная длина описания: %d

Нужно:
- 1 заголовок
- 1 описание
- акцент на пользе для клиента
- без ложных обещаний
- без эмодзи
- без HTML`,
		productName,
		in.ProductDescription,
		normalizeToneDescription(in.Tone),
		in.HeadlineMaxLen,
		in.BodyMaxLen,
	)

	return AIPromptSet{
		Messages: []AIPromptMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
	}
}

func BuildGenerateAdVariantsPrompt(in GenerateAdVariantsInput) AIPromptSet {
	in = normalizeGenerateAdVariantsInput(in)
	productName := strings.TrimSpace(in.ProductName)
	if productName == "" {
		productName = "не указано"
	}

	system := `Ты опытный рекламный копирайтер для A/B тестов. Твоя задача — создавать несколько разных рекламных вариантов на русском языке для товаров, услуг, заведений и digital-продуктов.

Правила:
- Пиши только на русском языке.
- Все варианты должны быть заметно разными по формулировке, но сохранять один и тот же смысл.
- Не выдумывай факты, скидки, цифры, гарантии или свойства, которых нет во входных данных.
- Не используй кликбейт, слишком громкие обещания и сомнительные рекламные формулировки.
- Каждый вариант должен состоять из headline и body.
- Заголовки должны быть короткими и разными по подаче.
- Описания должны быть ясными и ориентированными на выгоду.
- Если рекламируется физическое место, заведение или офлайн-услуга, допускай более атмосферную и жизненную подачу без выдуманных деталей.
- Если рекламируется товар или digital-продукт, делай варианты разными по углу подачи: выгода, удобство, результат, простота, экономия времени.
- Не добавляй комментарии, пояснения, markdown и ничего вне JSON.
- Верни результат строго в JSON-формате:
  {
    "variants": [
      { "headline": "строка", "body": "строка" }
    ]
  }
- Количество вариантов должно точно совпадать с запрошенным.
- Длина каждого headline не должна превышать ограничение пользователя.
- Длина каждого body не должна превышать ограничение пользователя.`

	user := fmt.Sprintf(`Сгенерируй %d A/B-вариантов рекламного объявления.

Данные о продукте:
Название продукта: %s
Описание товара или услуги: %s
Тон: %s

Ограничения:
Максимальная длина заголовка: %d
Максимальная длина описания: %d

Нужно:
- разные варианты подачи
- акцент на выгоде
- без ложных обещаний
- без эмодзи
- без HTML`,
		in.Count,
		productName,
		in.ProductDescription,
		normalizeToneDescription(in.Tone),
		in.HeadlineMaxLen,
		in.BodyMaxLen,
	)

	return AIPromptSet{
		Messages: []AIPromptMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
	}
}

func BuildGenerateAdImagePrompt(in GenerateAdImageInput) AIPromptSet {
	in = normalizeGenerateAdImageInput(in)

	system := `Ты создаёшь промпты для генерации рекламных изображений. Твоя задача — подготовить качественное описание визуала для интернет-рекламы.

Правила:
- Описание должно быть на английском языке (это улучшает качество генерации изображения).
- Изображение должно выглядеть как реальная премиальная рекламная фотография, а не как цифровой макет.
- Композиция должна быть чистой, понятной и визуально привлекательной.
- Не добавляй текст внутри изображения.
- Главный объект — реальный продукт, товар, место, еда или сцена из описания.
- Категорически избегай экранов, гаджетов, телефонов, ноутбуков, мониторов, интерфейсов, дашбордов, графиков, диаграмм, голограмм и футуристических хайтек-панелей, если пользователь явно не попросил об этом.
- Избегай запрещённого, шокирующего, сексуализированного, опасного или вводящего в заблуждение контента.
- Верни строго JSON:
  {
    "prompt": "строка"
  }`

	user := fmt.Sprintf(`Составь prompt для генерации рекламного изображения.

Описание продукта:
%s

Желаемый стиль:
%s

Формат объявления:
%s

Нужно:
- реальная премиальная рекламная фотография
- чистая композиция и сильный смысловой акцент
- без текста на изображении
- визуал должен передавать ценность продукта через реальный предмет, сцену или атмосферу
- если продукт физический — покажи сам товар как главный объект
- если продукт — услуга или цифровой сервис — покажи реальную сцену, людей или предметную композицию, передающую пользу, БЕЗ экранов и интерфейсов
- никаких устройств, экранов, интерфейсов, графиков и футуристических панелей`,
		in.Prompt,
		normalizeImageStyle(in.Style),
		normalizeAdFormatDescription(in.Format),
	)

	return AIPromptSet{
		Messages: []AIPromptMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
	}
}

// nanoBananaNegativePrompt — общий запрещающий блок. Главное лекарство от
// "устройств с экранами" и "футуристических графиков": нейробанан очень охотно
// рисует UI/дашборды/HUD, если их явно не запретить.
const nanoBananaNegativePrompt = "Strictly no text, no letters, no words, no numbers, no typography, no captions, no slogans, no logos, no watermark, no fake brand names, no UI labels, no poster or banner layout. " +
	"Strictly NO device screens, no smartphone, no tablet, no laptop, no computer, no monitor, no TV, no dashboard, no app interface, no website UI, no buttons, no holograms, no holographic panels, no floating screens, no charts, no graphs, no diagrams, no data visualizations, no infographics, no futuristic sci-fi HUD, no glowing tech interface, no abstract neon 3D shapes — unless the product description explicitly asks for them. " +
	"The result must look like a real, polished advertising photograph captured with a professional camera, not a digital mockup, app screenshot, 3D render, stock template or finished banner with copy."

func BuildNanoBananaImagePrompt(in GenerateAdImageInput) string {
	in = normalizeGenerateAdImageInput(in)

	formatInstruction := "16:9 horizontal advertising photo for feed placement"
	compositionInstruction := "leave clean breathing space, strong single focal point, balanced and natural composition"
	switch in.Format {
	case "stories":
		formatInstruction = "9:16 vertical advertising photo for stories placement"
		compositionInstruction = "vertical composition, subject centered or slightly lower, enough empty space near the top and bottom"
	}

	styleInstruction := normalizeImageStyle(in.Style)
	productDescription := strings.TrimSpace(in.Prompt)
	if productDescription == "" {
		productDescription = "the advertised product"
	}

	switch detectNanoBananaPromptMode(productDescription) {
	case nanoBananaPromptModeLifestyle:
		return fmt.Sprintf(
			"Create a high-quality, photorealistic commercial lifestyle advertising photo for %s. "+
				"Show the real place, environment, atmosphere, people, interior, furniture, food, drinks, packaging or moment described, "+
				"captured like authentic premium commercial photography: natural realistic lighting, believable real-world setting, "+
				"true-to-life colors, natural depth of field, inviting and emotionally warm. %s. %s. Style: %s. %s",
			productDescription,
			formatInstruction,
			compositionInstruction,
			styleInstruction,
			nanoBananaNegativePrompt,
		)
	default:
		return fmt.Sprintf(
			"Create a high-quality, photorealistic commercial advertising photo for %s. "+
				"If the product is a physical item, show the real product itself as the clear, beautifully lit hero subject on a tasteful real surface or background. "+
				"If the product is a service or digital product, show a relevant believable real-world scene, the people who benefit from it, or an elegant conceptual still life that conveys its value — never a screen, gadget or interface. "+
				"Make it look like premium professional advertising photography: realistic materials and textures, true-to-life colors, natural depth of field, clean and desirable. %s. %s. Style: %s. %s",
			productDescription,
			formatInstruction,
			compositionInstruction,
			styleInstruction,
			nanoBananaNegativePrompt,
		)
	}
}

type nanoBananaPromptMode string

const (
	// nanoBananaPromptModeProduct — дефолт: реальный товар/услуга как герой кадра,
	// без устройств и интерфейсов.
	nanoBananaPromptModeProduct   nanoBananaPromptMode = "product"
	nanoBananaPromptModeLifestyle nanoBananaPromptMode = "lifestyle"
)

func detectNanoBananaPromptMode(description string) nanoBananaPromptMode {
	normalized := strings.ToLower(strings.TrimSpace(description))
	if normalized == "" {
		return nanoBananaPromptModeProduct
	}

	lifestyleKeywords := []string{
		"кофей", "кафе", "кофе", "ресторан", "бар", "пекар", "булоч", "кондитер",
		"магазин", "бутик", "салон", "парикмах", "spa", "спа", "отел", "гостин",
		"терраса", "набереж", "пицц", "бургер", "суши", "еда", "десерт", "цветоч",
		"фитнес", "спортзал", "путешеств", "тур", "отдых", "клиник", "стоматолог",
		"coffee", "cafe", "restaurant", "bar", "bakery", "pastry", "dessert",
		"shop", "store", "salon", "hotel", "spa", "terrace", "embankment",
		"waterfront", "promenade", "pizza", "burger", "sushi", "flower", "boutique",
		"fitness", "gym", "travel", "tour", "clinic", "dental",
	}
	for _, keyword := range lifestyleKeywords {
		if strings.Contains(normalized, keyword) {
			return nanoBananaPromptModeLifestyle
		}
	}

	return nanoBananaPromptModeProduct
}

func normalizeToneDescription(tone string) string {
	switch strings.ToLower(strings.TrimSpace(tone)) {
	case "professional":
		return "деловой, уверенный, ясный"
	case "friendly":
		return "дружелюбный, тёплый, простой"
	case "bold":
		return "энергичный, смелый, но без агрессии"
	case "minimal":
		return "короткий, спокойный, чистый"
	case "":
		return "нейтральный, понятный, рекламный"
	default:
		return strings.TrimSpace(tone)
	}
}

// normalizeImageStyle переводит выбранный на фронте стиль ("Чистый" / "Яркий" /
// "Минимализм", либо их англоязычные ключи) в подробный визуальный дескриптор
// для генератора изображений.
func normalizeImageStyle(style string) string {
	switch strings.ToLower(strings.TrimSpace(style)) {
	case "", "чистый", "clean", "clean modern advertising visual":
		return "clean, fresh and premium commercial look; bright soft natural lighting; light, airy and uncluttered background; neutral refined color palette; crisp sharp focus; generous clean breathing space; modern, trustworthy, professional advertising photography"
	case "яркий", "bright", "vivid", "bold":
		return "vivid, bold and energetic commercial look; rich saturated colors with lively color accents; punchy high-contrast lighting; dynamic, eye-catching and joyful mood; expressive vibrant background; striking attention-grabbing premium advertising photography"
	case "минимализм", "минимал", "minimal", "minimalism", "minimalistic":
		return "minimalist commercial look; one clear hero subject; large amount of negative space; restrained muted or monochrome palette; simple plain background; calm, balanced and elegant composition; refined and uncluttered premium advertising photography"
	default:
		return strings.TrimSpace(style)
	}
}

func normalizeAdFormatDescription(format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "stories":
		return "stories, вертикальный 1080x1920"
	case "feed", "":
		return "лента, горизонтальный 1200x628"
	default:
		return format
	}
}
