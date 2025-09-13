# Project TODO List

## Done

-   Project scaffolding (`frontend`, `backend`, `parser` directories).
-   Backend Go module initialized.
-   Frontend Next.js application initialized.
-   Parser project files copied.
-   `docker-compose.yml` created for PostgreSQL, Redis, and Meilisearch.
-   Backend HTTP server created.
-   Backend `/api/search` endpoint created.
-   AI Normalizer Service (mocked) integrated into the backend.
-   Meilisearch integration (client, sample data indexing, search function) into the backend.
-   Meilisearch server upgraded to `last`.
-   Meilisearch data volume cleared to resolve compatibility issues.
-   Backend waits for Meilisearch indexing to complete before starting the HTTP server.
-   **AI Normalizer Service:**
    -   Replaced mock with actual Gemini API integration.
    -   Implemented key rotation for Gemini API keys.
    -   The prompt for the normalizer asks for title, category, and features.
    -   Handled the backticks in the Gemini API response.
-   **Meilisearch:**
    -   Implemented search for top-10 products by `title + features + category`.
    -   For each product, check `google_product_id`.
    -   Updated product in Meilisearch after fetching `google_product_id`.
    -   Changed Meilisearch hostname to `meilisearch` in `backend/internal/search/search.go`.
    -   Added `GetProduct` function to retrieve product by ID from Meilisearch.
-   **SerpApi Integration:**
    -   Implemented key rotation for SerpApi keys.
    -   If `google_product_id` is not present, form `serpapi_query` by title.
    -   Get the first result (`position=1`) from SerpApi.
    -   Extract and save `google_product_id`.
    -   Implemented `GetProductOffers` function to get merchants, prices, and links using `google_product_id`.
-   **Redis Cache:**
    -   Implemented caching for SerpApi results (merchants, prices, links) for 1 day.
    -   Changed Redis hostname to `redis` in `backend/internal/cache/redis.go`.
-   **Docker Compose:**
    -   Added `backend` and `frontend` services to `docker-compose.yml`.
    -   Added `Dockerfile` for `backend` and `frontend`.
    -   Added healthcheck for `meilisearch` service.
    -   Updated Go version in `backend/Dockerfile`.
-   **Backend:**
    -   Fixed issue where backend was not accessible from the host.
-   **Frontend:**
    -   Design: like ChatGPT.
    -   Automatic language detection.
    -   Automatic region detection.
    -   Query formation with lang and region.
    -   Display top-10 products as a list with image and name.
    -   **User Clicks on Product:**
        -   Display detailed product card (image, name, specification, "Show prices and stores" button) as a separate page.
        -   On "Show prices and stores" button click, display merchants, prices, and links.
        -   Implement swipe-right to go back to product list.
-   **Parser PriceRunner:**
    -   Improved the parser to be manageable via admin panel (start, stop, status).
    -   Implemented daily parsing for price updates.
    -   The parser updates PostgreSQL and Meilisearch with new data.
-   **Admin Panel (Frontend Next.js + Go Backend API):**
    -   Manage API keys for SerpApi and AI Normalizer Service Gemini.
    -   View request logs and errors.
    -   Manage products in the database (add, delete, edit).
    -   View usage statistics (number of requests, popular products, etc.).

## To Do (from ShemaProject.ini)

Проверь весь проект где можно улучшить и улучшы Frontend исходя из задачи в ./ShemaProject.ini
Не пытайся собрать или запустить проект только реализация кода