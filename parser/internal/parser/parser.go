package parser

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"pricerunner-parser/internal/config"
	"pricerunner-parser/internal/downloader"
	"pricerunner-parser/internal/models"
	"pricerunner-parser/internal/storage"

	"github.com/playwright-community/playwright-go"
)

// Parser handles the main parsing logic
type Parser struct {
	config     *config.Config
	storage    storage.Storage
	downloader *downloader.ImageDownloader
	playwright *playwright.Playwright
	browser    playwright.Browser
}

// New creates a new parser instance
func New(cfg *config.Config, store storage.Storage) *Parser {
	return &Parser{
		config:     cfg,
		storage:    store,
		downloader: downloader.NewImageDownloader(cfg.Storage.ImagesDir),
	}
}

// Parse starts the parsing process
func (p *Parser) Parse() error {
	// Инициализируем Playwright
	if err := p.initPlaywright(); err != nil {
		return fmt.Errorf("failed to init playwright: %w", err)
	}
	defer p.cleanup()

	var allProducts []models.Product
	pageNumber := 1
	maxPages := p.config.Parser.Parsing.MaxPages

	// Если maxPages = 0, ставим разумный лимит чтобы не уходить в бесконечность
	if maxPages == 0 {
		maxPages = 50 // Разумный лимит
	}

	for pageNumber <= maxPages {
		log.Printf("\n=== Processing page %d ===", pageNumber)

		// Парсим список товаров на странице
		basicProducts, hasNextPage, err := p.parseProductList(pageNumber)
		if err != nil {
			log.Printf("Failed to parse page %d: %v", pageNumber, err)
			break
		}

		if len(basicProducts) == 0 {
			log.Printf("No products found on page %d, stopping", pageNumber)
			break
		}

		log.Printf("Found %d products on page %d", len(basicProducts), pageNumber)

		// Фильтруем уже существующие товары
		newProducts, err := p.filterExistingProducts(basicProducts)
		if err != nil {
			log.Printf("Warning: failed to filter existing products: %v", err)
			newProducts = basicProducts // Продолжаем со всеми товарами
		}

		log.Printf("New products to process: %d", len(newProducts))

		// Получаем детальную информацию
		detailedProducts := p.parseProductDetails(newProducts)

		// Сохраняем данные страницы
		if len(detailedProducts) > 0 {
			if err := p.storage.SaveProducts(detailedProducts, pageNumber); err != nil {
				log.Printf("Warning: failed to save page %d: %v", pageNumber, err)
			}
			allProducts = append(allProducts, detailedProducts...)
		}

		log.Printf("=== Page %d completed: %d products processed ===", pageNumber, len(detailedProducts))

		// Проверяем нужно ли продолжать
		if !hasNextPage {
			log.Printf("No more pages available, stopping at page %d", pageNumber)
			break
		}

		// Переходим к следующей странице
		pageNumber++
		log.Printf("Moving to page %d...", pageNumber)
		time.Sleep(time.Duration(p.config.Parser.Parsing.DelayBetweenRequests) * time.Millisecond)
	}

	// Сохраняем финальные данные
	if len(allProducts) > 0 {
		if err := p.storage.SaveFinalData(allProducts); err != nil {
			return fmt.Errorf("failed to save final data: %w", err)
		}

		log.Printf("\n🎉 PARSING COMPLETED SUCCESSFULLY!")
		log.Printf("📊 Total processed: %d products across %d pages", len(allProducts), pageNumber-1)

		// Статистика по компонентам
		withPrices := 0
		withImages := 0
		withFeatures := 0

		for _, product := range allProducts {
			if product.Price != nil && product.Price.PriceEUR > 0 {
				withPrices++
			}
			if product.ImageLocal != "" {
				withImages++
			}
			if len(product.Features) > 0 {
				withFeatures++
			}
		}

		log.Printf("💰 Products with prices: %d/%d", withPrices, len(allProducts))
		log.Printf("🖼️  Products with images: %d/%d", withImages, len(allProducts))
		log.Printf("⚙️  Products with features: %d/%d", withFeatures, len(allProducts))
	}

	return nil
}

// initPlaywright initializes Playwright browser
func (p *Parser) initPlaywright() error {
	pw, err := playwright.Run()
	if err != nil {
		return err
	}
	p.playwright = pw

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: &p.config.Parser.Browser.Headless,
		Args: []string{
			"--disable-blink-features=AutomationControlled",
			"--no-sandbox",
			"--disable-setuid-sandbox",
		},
	})
	if err != nil {
		return err
	}
	p.browser = browser

	return nil
}

// cleanup closes browser and playwright
func (p *Parser) cleanup() {
	if p.browser != nil {
		p.browser.Close()
	}
	if p.playwright != nil {
		p.playwright.Stop()
	}
}

// parseProductList parses the product list from a page
func (p *Parser) parseProductList(pageNumber int) ([]models.BasicProduct, bool, error) {
	context, err := p.browser.NewContext(playwright.BrowserNewContextOptions{
		Viewport: &playwright.Size{
			Width:  p.config.Parser.Browser.Viewport.Width,
			Height: p.config.Parser.Browser.Viewport.Height,
		},
		UserAgent: &p.config.Parser.Browser.UserAgent,
	})
	if err != nil {
		return nil, false, err
	}
	defer context.Close()

	page, err := context.NewPage()
	if err != nil {
		return nil, false, err
	}

	// Переходим на страницу с номером
	url := fmt.Sprintf("%s?page=%d", p.config.Parser.BaseURL, pageNumber)
	log.Printf("Loading page: %s", url)

	response, err := page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		Timeout:   playwright.Float(float64(p.config.Parser.Browser.Timeout)),
	})
	if err != nil {
		return nil, false, err
	}

	// Проверяем статус ответа
	if response.Status() == 404 {
		log.Printf("  Page %d returned 404, no more pages", pageNumber)
		return nil, false, nil
	}

	// Ждем загрузки товаров
	_, err = page.WaitForSelector(p.config.Parser.Selectors.ProductCards, playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(10000),
	})
	if err != nil {
		log.Printf("  No products found on page %d", pageNumber)
		return nil, false, nil
	}

	// Прокручиваем страницу для загрузки всех товаров
	p.scrollPage(page)

	// Извлекаем товары
	products := p.extractProductCards(page)

	if len(products) == 0 {
		log.Printf("  No products extracted from page %d", pageNumber)
		return nil, false, nil
	}

	// Определяем есть ли следующая страница по количеству товаров
	hasNextPage := p.hasNextPage(len(products), pageNumber)

	return products, hasNextPage, nil
}

// scrollPage scrolls the page to load all products
func (p *Parser) scrollPage(page playwright.Page) {
	scrollDelay := time.Duration(p.config.Parser.Parsing.ScrollDelay) * time.Millisecond
	maxScrolls := p.config.Parser.Parsing.MaxScrolls

	log.Println("Scrolling page to load all products...")

	var lastHeight float64
	stableCount := 0

	for i := 0; i < maxScrolls; i++ {
		// Прокручиваем вниз
		page.Evaluate("window.scrollBy(0, 500)")
		time.Sleep(scrollDelay)

		// Проверяем изменилась ли высота
		heightResult, err := page.Evaluate("document.body.scrollHeight")
		if err != nil {
			break
		}

		var currentHeight float64
		switch h := heightResult.(type) {
		case int:
			currentHeight = float64(h)
		case int64:
			currentHeight = float64(h)
		case float64:
			currentHeight = h
		case float32:
			currentHeight = float64(h)
		default:
			currentHeight = 0
		}

		if currentHeight == lastHeight {
			stableCount++
			if stableCount >= 3 {
				break
			}
		} else {
			stableCount = 0
		}

		lastHeight = currentHeight
	}

	// Возвращаемся наверх
	page.Evaluate("window.scrollTo(0, 0)")
	time.Sleep(time.Second)
}

// extractProductCards extracts product information from page
func (p *Parser) extractProductCards(page playwright.Page) []models.BasicProduct {
	cards, err := page.QuerySelectorAll(p.config.Parser.Selectors.ProductCards)
	if err != nil {
		log.Printf("Error querying product cards: %v", err)
		return nil
	}

	var products []models.BasicProduct

	for _, card := range cards {
		title, err := card.GetAttribute("title")
		if err != nil || title == "" {
			continue
		}

		href, err := card.GetAttribute("href")
		if err != nil || href == "" {
			continue
		}

		// Извлекаем ID из URL
		re := regexp.MustCompile(`/pl/(\d+-\d+)/`)
		matches := re.FindStringSubmatch(href)
		if len(matches) < 2 {
			continue
		}

		productID := matches[1]

		// Формируем полный URL
		fullURL := href
		if !strings.HasPrefix(href, "http") {
			fullURL = "https://www.pricerunner.com" + href
		}

		products = append(products, models.BasicProduct{
			ID:    productID,
			Title: title,
			URL:   fullURL,
		})
	}

	return products
}

// hasNextPage checks if there's a next page based on product count
func (p *Parser) hasNextPage(productsCount int, pageNumber int) bool {
	log.Printf("  Checking if page %d has next page...", pageNumber)

	// Если на странице меньше товаров чем обычно, скорее всего это последняя страница
	expectedProductsPerPage := 48

	if productsCount < expectedProductsPerPage {
		log.Printf("    Found only %d products (expected ~%d), likely last page", productsCount, expectedProductsPerPage)
		return false
	}

	// Если достигли максимального числа страниц из конфига
	if p.config.Parser.Parsing.MaxPages > 0 && pageNumber >= p.config.Parser.Parsing.MaxPages {
		log.Printf("    Reached max pages limit (%d)", p.config.Parser.Parsing.MaxPages)
		return false
	}

	log.Printf("    ✓ Page %d has %d products, assuming next page exists", pageNumber, productsCount)
	return true
}

// filterExistingProducts filters out products that already exist
func (p *Parser) filterExistingProducts(products []models.BasicProduct) ([]models.BasicProduct, error) {
	existingIDs, err := p.storage.GetExistingProducts()
	if err != nil {
		return products, err // Возвращаем все товары если не можем проверить
	}

	existingSet := make(map[string]bool)
	for _, id := range existingIDs {
		existingSet[id] = true
	}

	var newProducts []models.BasicProduct
	for _, product := range products {
		if !existingSet[product.ID] {
			newProducts = append(newProducts, product)
		}
	}

	return newProducts, nil
}

// parseProductDetails gets detailed information for each product
func (p *Parser) parseProductDetails(products []models.BasicProduct) []models.Product {
	var detailedProducts []models.Product

	for i, basicProduct := range products {
		log.Printf("[%d/%d] Processing: %s (ID: %s)", i+1, len(products),
			truncateString(basicProduct.Title, 50), basicProduct.ID)

		details, err := p.parseProductDetail(basicProduct)
		if err != nil {
			log.Printf("  ✗ Failed to parse details: %v", err)
			continue
		}

		detailedProducts = append(detailedProducts, *details)

		// Компактная статистика
		var stats []string

		if details.Price != nil && details.Price.PriceEUR > 0 {
			stats = append(stats, fmt.Sprintf("Price: €%.2f", details.Price.PriceEUR))
		}

		if details.ImageLocal != "" {
			stats = append(stats, "Image: ✓")
		}

		if len(details.Features) > 0 {
			stats = append(stats, fmt.Sprintf("Features: %d", len(details.Features)))
		}

		if len(details.ExtraImages) > 0 {
			stats = append(stats, fmt.Sprintf("ExtraImg: %d", len(details.ExtraImages)))
		}

		// Выводим все в одной строке
		if len(stats) > 0 {
			log.Printf("  ✓ %s", strings.Join(stats, ", "))
		}

		// Пауза между запросами
		time.Sleep(time.Duration(p.config.Parser.Parsing.DelayBetweenRequests) * time.Millisecond)
	}

	return detailedProducts
}

// parseProductDetail parses detailed information for a single product
func (p *Parser) parseProductDetail(basic models.BasicProduct) (*models.Product, error) {
	context, err := p.browser.NewContext(playwright.BrowserNewContextOptions{
		Viewport: &playwright.Size{
			Width:  p.config.Parser.Browser.Viewport.Width,
			Height: p.config.Parser.Browser.Viewport.Height,
		},
		UserAgent: &p.config.Parser.Browser.UserAgent,
	})
	if err != nil {
		return nil, err
	}
	defer context.Close()

	page, err := context.NewPage()
	if err != nil {
		return nil, err
	}

	// Переходим на страницу товара
	if _, err := page.Goto(basic.URL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		Timeout:   playwright.Float(float64(p.config.Parser.Browser.Timeout)),
	}); err != nil {
		return nil, err
	}

	// Минимальная пауза для начальной загрузки
	time.Sleep(500 * time.Millisecond)

	// Универсальная прокрутка для загрузки всего контента
	p.scrollToFeatures(page)

	// Создаем объект продукта
	product := &models.Product{
		ID:        basic.ID,
		Title:     basic.Title,
		URL:       basic.URL,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Парсим все данные
	product.Price = p.parsePrice(page)
	p.parseMainImage(page, product)
	product.ExtraImages = p.parseAdditionalImages(page, basic.ID)
	product.Features = p.parseFeatures(page)

	return product, nil
}

// scrollToFeatures улучшенная прокрутка для загрузки всего контента
func (p *Parser) scrollToFeatures(page playwright.Page) {
	log.Printf("  Loading all page content...")

	// Стратегия: быстро прокрутить всю страницу для загрузки контента

	// 1. Прокручиваем до конца страницы
	page.Evaluate("window.scrollTo(0, document.body.scrollHeight)")
	time.Sleep(1 * time.Second)

	// 2. Проверяем есть ли таблицы
	tablesCount := p.checkTablesCountQuick(page)
	log.Printf("    After full scroll: %d tables found", tablesCount)

	// 3. Если таблиц мало, делаем дополнительные прокрутки
	if tablesCount < 2 {
		log.Printf("    Need more content, doing additional scrolling...")

		// Прокручиваем к разным частям страницы
		positions := []string{
			"document.body.scrollHeight * 0.75", // 75%
			"document.body.scrollHeight * 0.5",  // 50%
			"document.body.scrollHeight * 0.25", // 25%
		}

		for _, pos := range positions {
			page.Evaluate(fmt.Sprintf("window.scrollTo(0, %s)", pos))
			time.Sleep(500 * time.Millisecond)
		}

		// Финальная прокрутка к концу
		page.Evaluate("window.scrollTo(0, document.body.scrollHeight)")
		time.Sleep(1 * time.Second)
	}

	// 4. Позиционируемся в середине для парсинга
	page.Evaluate("window.scrollTo(0, document.body.scrollHeight / 2)")
	time.Sleep(500 * time.Millisecond)

	finalCount := p.checkTablesCountQuick(page)
	log.Printf("    Final: %d tables loaded", finalCount)
}

// checkTablesCountQuick быстрая проверка количества таблиц
func (p *Parser) checkTablesCountQuick(page playwright.Page) int {
	// Пробуем основной селектор
	tables, err := page.QuerySelectorAll(p.config.Parser.Selectors.FeatureTables)
	if err == nil && len(tables) > 0 {
		return len(tables)
	}

	// Пробуем альтернативные селекторы
	altSelectors := []string{
		"table.pr-1regpt0-Table-table",
		"div[class*='Table'] table",
		"table",
	}

	maxCount := 0
	for _, selector := range altSelectors {
		tables, err := page.QuerySelectorAll(selector)
		if err == nil && len(tables) > maxCount {
			maxCount = len(tables)
		}
	}

	return maxCount
}

// parsePrice parses price using the improved selector for second span
func (p *Parser) parsePrice(page playwright.Page) *models.PriceInfo {
	log.Printf("    Looking for price (second span)...")

	// Способ 1: Ищем все span'ы с нужным классом и берем второй
	priceElements, err := page.QuerySelectorAll("span.pr-1fcg5be")
	if err != nil || len(priceElements) < 2 {
		log.Printf("    Not enough price elements found (need 2, got %d)", len(priceElements))
		return p.parsePriceFallback(page)
	}

	// Берем второй элемент (индекс 1)
	secondPriceElement := priceElements[1]

	priceText, err := secondPriceElement.InnerText()
	if err != nil || priceText == "" {
		log.Printf("    Failed to get text from second price element")
		return p.parsePriceFallback(page)
	}

	priceText = strings.TrimSpace(priceText)
	log.Printf("    Found price (second span): %s", priceText)

	// Конвертируем в EUR
	priceEUR := p.convertGBPToEUR(priceText)

	return &models.PriceInfo{
		PriceGBP: priceText,
		PriceEUR: priceEUR,
	}
}

// parsePriceFallback - альтернативные способы поиска цены
func (p *Parser) parsePriceFallback(page playwright.Page) *models.PriceInfo {
	log.Printf("    Trying fallback price selectors...")

	// Способ 2: Ищем по тексту "Lowest Price Now"
	nowPriceSelector := `//p[contains(text(), "Lowest Price Now")]/following-sibling::span[@class="pr-1fcg5be"]`
	if priceElement, err := page.QuerySelector(nowPriceSelector); err == nil && priceElement != nil {
		if priceText, err := priceElement.InnerText(); err == nil && priceText != "" {
			log.Printf("    Found price via 'Lowest Price Now': %s", strings.TrimSpace(priceText))
			return &models.PriceInfo{
				PriceGBP: strings.TrimSpace(priceText),
				PriceEUR: p.convertGBPToEUR(priceText),
			}
		}
	}

	// Способ 3: Ищем по структуре DOM - второй div с классом pr-i5pc8s
	structureSelector := `div.pr-1ymxntz div.pr-i5pc8s:nth-child(2) span.pr-1fcg5be`
	if priceElement, err := page.QuerySelector(structureSelector); err == nil && priceElement != nil {
		if priceText, err := priceElement.InnerText(); err == nil && priceText != "" {
			log.Printf("    Found price via structure selector: %s", strings.TrimSpace(priceText))
			return &models.PriceInfo{
				PriceGBP: strings.TrimSpace(priceText),
				PriceEUR: p.convertGBPToEUR(priceText),
			}
		}
	}

	log.Printf("    ⚠ No price found with any method")
	return nil
}

// convertGBPToEUR converts GBP price to EUR
func (p *Parser) convertGBPToEUR(priceStr string) float64 {
	// Удаляем символы валюты и пробелы
	re := regexp.MustCompile(`[£€\s,]`)
	cleanPrice := re.ReplaceAllString(priceStr, "")

	price, err := strconv.ParseFloat(cleanPrice, 64)
	if err != nil {
		return 0
	}

	return price * p.config.Currency.GBPToEUR
}

// parseMainImage parses the main product image
func (p *Parser) parseMainImage(page playwright.Page, product *models.Product) {
	log.Printf("  Looking for images...")

	selectors := []string{
		p.config.Parser.Selectors.MainImage,
		"picture.pr-lpjxdi source[type='image/jpeg']",
		"img[itemprop='image']",
		"div.pr-15dcama img",
	}

	var imageURL string

	for _, selector := range selectors {
		if strings.Contains(selector, "source") {
			source, err := page.QuerySelector(selector)
			if err == nil && source != nil {
				srcset, _ := source.GetAttribute("srcset")
				if srcset != "" {
					urls := strings.Split(srcset, ",")
					if len(urls) > 0 {
						imageURL = strings.TrimSpace(strings.Fields(urls[len(urls)-1])[0])
						break
					}
				}
			}
		} else {
			img, err := page.QuerySelector(selector)
			if err == nil && img != nil {
				srcset, _ := img.GetAttribute("srcset")
				if srcset != "" {
					urls := strings.Split(srcset, ",")
					if len(urls) > 0 {
						imageURL = strings.TrimSpace(strings.Fields(urls[len(urls)-1])[0])
						break
					}
				} else {
					src, _ := img.GetAttribute("src")
					if src != "" && !strings.HasPrefix(src, "data:") {
						imageURL = src
						break
					}
				}
			}
		}
	}

	product.ImageURL = imageURL

	if imageURL != "" {
		localPath, err := p.downloader.DownloadImage(imageURL, product.ID)
		if err != nil {
			log.Printf("    ✗ Failed to download main image: %v", err)
		} else {
			product.ImageLocal = localPath
		}
	} else {
		log.Printf("    ⚠ Main image not found")
	}
}

// parseAdditionalImages parses additional product images
func (p *Parser) parseAdditionalImages(page playwright.Page, productID string) []models.ImageInfo {
	log.Printf("  Looking for additional images...")

	thumbnails, err := page.QuerySelectorAll(p.config.Parser.Selectors.AdditionalImages)
	if err != nil {
		return nil
	}

	var additionalImages []models.ImageInfo
	maxImages := 3 // Ограничиваем для скорости

	for i, thumb := range thumbnails {
		if i >= maxImages {
			break
		}

		src, err := thumb.GetAttribute("src")
		if err != nil || src == "" || strings.HasPrefix(src, "data:") {
			continue
		}

		// Заменяем размер на больший
		re := regexp.MustCompile(`/dim/dim/`)
		largeURL := re.ReplaceAllString(src, "/504x504/")

		localPath, err := p.downloader.DownloadImage(largeURL, fmt.Sprintf("%s_extra_%d", productID, i+1))
		if err != nil {
			log.Printf("    ✗ Failed to download additional image %d: %v", i+1, err)
			continue
		}

		additionalImages = append(additionalImages, models.ImageInfo{
			URL:   largeURL,
			Local: localPath,
		})
	}

	return additionalImages
}

// parseFeatures парсинг таблицы с множественными заголовками внутри
func (p *Parser) parseFeatures(page playwright.Page) map[string]string {
	log.Printf("  Extracting features...")

	features := make(map[string]string)

	// Пробуем разные селекторы для таблиц
	tableSelectors := []string{
		p.config.Parser.Selectors.FeatureTables, // Основной
		"table.pr-1regpt0-Table-table",          // Альтернативный
		"div[class*='Table'] table",             // По части класса
		"table",                                 // Любые таблицы
	}

	var tables []playwright.ElementHandle

	for _, selector := range tableSelectors {
		foundTables, err := page.QuerySelectorAll(selector)
		if err == nil && len(foundTables) > 0 {
			tables = foundTables
			log.Printf("    Found %d tables with selector: %s", len(foundTables), selector)
			break
		}
	}

	if len(tables) == 0 {
		log.Printf("    ⚠ No feature tables found with any selector")
		return features
	}

	// Парсим каждую таблицу
	for tableIdx, table := range tables {
		log.Printf("    Processing table %d/%d", tableIdx+1, len(tables))

		// Парсим таблицу с множественными заголовками
		tableFeatures := p.parseTableWithMultipleHeaders(table)

		// Объединяем с общим списком
		for key, value := range tableFeatures {
			features[key] = value
		}
	}

	if len(features) > 0 {
		log.Printf("    ✓ Total features found: %d", len(features))
	} else {
		log.Printf("    ⚠ No valid features extracted")
	}

	return features
}

// Исправляем parseTableWithMultipleHeaders - убираем неиспользуемую переменную
func (p *Parser) parseTableWithMultipleHeaders(table playwright.ElementHandle) map[string]string {
	features := make(map[string]string)

	// Получаем все строки таблицы (и thead, и tbody)
	allRows, err := table.QuerySelectorAll("tr")
	if err != nil {
		return features
	}

	currentCategory := "General"

	for _, row := range allRows { // Убрали i, _
		// Проверяем является ли строка заголовком
		if p.isHeaderRow(row) {
			// Это строка заголовка - извлекаем категорию
			category := p.extractCategoryFromHeaderRow(row)
			if category != "" {
				currentCategory = category
				log.Printf("      Found category header: '%s'", currentCategory)
			}
			continue
		}

		// Это строка данных - парсим характеристики
		featureName, featureValue := p.extractFeatureFromRow(row)
		if featureName != "" && featureValue != "" {
			// Добавляем категорию к названию если она значимая
			key := featureName
			if currentCategory != "General" && currentCategory != "" {
				key = fmt.Sprintf("%s: %s", currentCategory, featureName)
			}

			features[key] = featureValue
		}
	}

	log.Printf("      Extracted %d features from table", len(features))
	return features
}

// isHeaderRow проверяет является ли строка заголовком
func (p *Parser) isHeaderRow(row playwright.ElementHandle) bool {
	// Способ 1: Проверяем родительский элемент (thead)
	script := "\n\t\t\t\tlet row = arguments[0];\n\t\t\t\tlet parent = row.parentElement;\n\t\t\t\treturn parent && parent.tagName === 'THEAD';\n\t\t"
	if result, err := row.Evaluate(script); err == nil {
		if isInThead, ok := result.(bool); ok && isInThead {
			return true
		}
	}

	// Способ 2: Проверяем наличие th вместо td
	ths, err := row.QuerySelectorAll("th")
	if err == nil && len(ths) > 0 {
		return true
	}

	// Способ 3: Проверяем CSS класс или data-атрибуты
	class, _ := row.GetAttribute("class")
	dataKind, _ := row.GetAttribute("data-kind")

	if strings.Contains(class, "heading") || strings.Contains(class, "header") ||
		dataKind == "heading" || dataKind == "header" {
		return true
	}

	// Способ 4: Проверяем ячейки с data-kind="heading"
	cells, err := row.QuerySelectorAll("td, th")
	if err == nil {
		for _, cell := range cells {
			if cellDataKind, _ := cell.GetAttribute("data-kind"); cellDataKind == "heading" {
				return true
			}
		}
	}

	return false
}

// extractCategoryFromHeaderRow извлекает название категории из строки заголовка
func (p *Parser) extractCategoryFromHeaderRow(row playwright.ElementHandle) string {
	// Ищем div.pr-f8aw3g с текстом
	if categoryDiv, err := row.QuerySelector("div.pr-f8aw3g"); err == nil && categoryDiv != nil {
		if categoryText, err := categoryDiv.InnerText(); err == nil && categoryText != "" {
			cleaned := strings.TrimSpace(categoryText)
			if p.isValidCategory(cleaned) {
				return cleaned
			}
		}
	}

	// Альтернативно - ищем в th
	if th, err := row.QuerySelector("th"); err == nil && th != nil {
		if categoryText, err := th.InnerText(); err == nil && categoryText != "" {
			cleaned := strings.TrimSpace(categoryText)
			if p.isValidCategory(cleaned) {
				return cleaned
			}
		}
	}

	// Ищем в любой ячейке
	cells, err := row.QuerySelectorAll("td, th")
	if err == nil && len(cells) > 0 {
		if categoryText, err := cells[0].InnerText(); err == nil && categoryText != "" {
			cleaned := strings.TrimSpace(categoryText)
			if p.isValidCategory(cleaned) {
				return cleaned
			}
		}
	}

	return ""
}

// extractFeatureFromRow извлекает характеристику из строки данных
func (p *Parser) extractFeatureFromRow(row playwright.ElementHandle) (string, string) {
	// Получаем ячейки td (не th)
	cells, err := row.QuerySelectorAll("td")
	if err != nil || len(cells) < 2 {
		return "", ""
	}

	// Извлекаем название и значение
	featureName, _ := cells[0].InnerText()
	featureValue, _ := cells[1].InnerText()

	featureName = strings.TrimSpace(featureName)
	featureValue = strings.TrimSpace(featureValue)

	// Валидация
	if !p.isValidFeaturePair(featureName, featureValue) {
		return "", ""
	}

	return featureName, featureValue
}

// isValidCategory проверяет является ли строка валидной категорией
func (p *Parser) isValidCategory(category string) bool {
	if len(category) < 2 || len(category) > 50 {
		return false
	}

	// Исключаем пустые или служебные категории
	excludes := []string{",", " ", "-", "–", "—", "N/A", "n/a", "TBD", "tbd"}
	for _, exclude := range excludes {
		if category == exclude {
			return false
		}
	}

	return true
}

// isValidFeaturePair проверяет валидность пары название-значение
func (p *Parser) isValidFeaturePair(name, value string) bool {
	if name == "" || value == "" {
		return false
	}

	if name == value {
		return false
	}

	if len(name) < 2 || len(value) < 1 {
		return false
	}

	// Исключаем служебные строки
	if strings.Contains(strings.ToLower(name), "compare") {
		return false
	}

	return true
}

// Добавляем недостающую функцию truncateString в конец файла
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
