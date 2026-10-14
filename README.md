# Go Flights Daemon

A Go-based HTTP API that loads flight data from multiple CSV files, stores flights in an in-memory cache, and provides filtering, sorting, and pagination through a JSON API.

## Features

- **Concurrent data loading:** Reads flight data from multiple CSV files concurrently.
- **In-memory caching:** Stores parsed flights in a cache indexed by departure airport, destination airport, departure date, and return date.
- **Flight filtering:** Filters flights by departure airport, destination airport, departure date, and optionally return date and airlines.
- **Price sorting:** Sorts flights by price in ascending or descending order.
- **Pagination:** Supports paginated responses to control the number of flights returned per request.
- **JSON responses:** Returns flight data and pagination information as JSON.
- **Partial data availability:** The API can start serving requests while flight data is still loading. Responses may therefore omit flights that have not yet been loaded.
- **Docker support:** Can be built and run in a Docker container.

## API

### `GET /flights`

Returns flights matching the specified search criteria as a JSON response.

#### Query Parameters

| Parameter | Required | Description |
|---|---|---|
| `departure` | Yes | Departure airport code. |
| `destination` | Yes | Destination airport code. |
| `departure_date` | Yes | Departure date. |
| `return_date` | No | Return date. |
| `airlines[]` | No | Filters results by one or more airline codes. |
| `page` | No | Page number. |
| `per_page` | No | Number of flights per page. |
| `sort_direction` | No | Sort direction based on flight price: `ASC` or `DESC`. |

#### Sorting

Flights are sorted by price in ascending order by default.

- `ASC`: Lowest price first.
- `DESC`: Highest price first.

If `sort_direction` is omitted, `ASC` is used.

Only `ASC` and `DESC` are supported.

#### Pagination

The API supports pagination through the `page` and `per_page` parameters. These parameters are optional.

**Request example:**

`http://localhost:8080/flights?departure=CGN&destination=STR&departure_date=2022-06-10&return_date=2022-06-16&airlines%5B%5D=LH&airlines%5B%5D=2A&page=1&per_page=2&sort_direction=ASC`

#### Response

The endpoint returns a JSON response containing the matching flights and pagination information.

**Example response:**

```json
{
  "runtime": 0.000046375,
  "flight_offers": 60,
  "page_count": 30,
  "current_page": 1,
  "per_page": 2,
  "airline_offers": [
    {
      "code": "LH",
      "price_min": 234.25,
      "price_max": 244.31
    }
  ],
  "flights": [
    {
      "departure": "CGN",
      "destination": "STR",
      "date_departure": "2022-06-10",
      "date_return": "2022-06-10",
      "carrier": "LH",
      "price": 234.25
    },
    {
      "departure": "CGN",
      "destination": "STR",
      "date_departure": "2022-06-10",
      "date_return": "2022-06-10",
      "carrier": "LH",
      "price": 234.25
    }
  ]
}
```

## Data Loading and Availability

At startup, the daemon reads flight data from multiple CSV files and stores the parsed records in its in-memory cache.

Data loading is performed concurrently, allowing the API to start serving requests without waiting for every file to finish loading.

As a result, responses may contain only a subset of the matching flights while loading is in progress. Flights that have not yet been loaded into the cache will not appear in the response.

## Running with Docker Compose

The project uses Docker Compose to build and run the flights daemon. The CSV files required by the application are included in the repository.

Start the application:

```bash
docker compose up
```

Once the application is running, the API is available at the port configured in `docker-compose.yml`.

## Technology

- **Go** — HTTP API and application logic
- **CSV** — Flight data ingestion using `^`-delimited records
- **In-memory cache** — Fast retrieval of flights by search criteria
- **Docker** — Containerized application deployment