package main

import (
	"fmt"
	"math"
	"sync"
)

// City структура для представлення міста в графі
type City struct {
	name string
}

// Edge структура для представлення рейсу між містами та ціни квитка
type Edge struct {
	from            *City
	to              *City
	price           int
	isBidirectional bool // Чи рейс є обидвостороннім (від А до Б та від Б до А)
}

// Graph структура для представлення графу міст та рейсів
type Graph struct {
	cities []*City
	edges  []*Edge
	lock   sync.RWMutex
}

// NewGraph створення нового графу
func NewGraph() *Graph {
	return &Graph{}
}

// AddCity додавання нового міста до графу
func (g *Graph) AddCity(name string) *City {
	g.lock.Lock()
	defer g.lock.Unlock()

	city := &City{name: name}
	g.cities = append(g.cities, city)
	return city
}

// AddEdge додавання нового рейсу між містами та ціни квитка
func (g *Graph) AddEdge(from, to *City, price int, isBidirectional bool) {
	g.lock.Lock()
	defer g.lock.Unlock()

	edge := &Edge{from: from, to: to, price: price, isBidirectional: isBidirectional}
	g.edges = append(g.edges, edge)

	// Якщо рейс є обидвостороннім, додати ще один рейс у зворотному напрямку
	if isBidirectional {
		reverseEdge := &Edge{from: to, to: from, price: price, isBidirectional: isBidirectional}
		g.edges = append(g.edges, reverseEdge)
	}
}

// RemoveEdge видалення рейсу між містами
func (g *Graph) RemoveEdge(from, to *City) {
	g.lock.Lock()
	defer g.lock.Unlock()

	for i, edge := range g.edges {
		if edge.from == from && edge.to == to {
			g.edges = append(g.edges[:i], g.edges[i+1:]...)
			break
		}
	}
}

// ChangeTicketPrice зміна ціни квитка для рейсу між містами
func (g *Graph) ChangeTicketPrice(from, to *City, newPrice int) {
	g.lock.Lock()
	defer g.lock.Unlock()

	for _, edge := range g.edges {
		if edge.from == from && edge.to == to {
			edge.price = newPrice
			break
		}
	}
}

// RemoveCity видалення міста з графу
func (g *Graph) RemoveCity(city *City) {
	g.lock.Lock()
	defer g.lock.Unlock()

	// Видалити рейси, пов'язані з цим містом
	var updatedEdges []*Edge
	for _, edge := range g.edges {
		if edge.from != city && edge.to != city {
			updatedEdges = append(updatedEdges, edge)
		}
	}
	g.edges = updatedEdges

	// Видалити саме місто
	var updatedCities []*City
	for _, c := range g.cities {
		if c != city {
			updatedCities = append(updatedCities, c)
		}
	}
	g.cities = updatedCities
}

// FindPathAndPrice знаходження шляху та ціни квитка від міста A до міста B
func (g *Graph) FindPathAndPrice(start, end *City) (path []*City, totalPrice int) {
	g.lock.RLock()
	defer g.lock.RUnlock()

	visited := make(map[*City]bool)
	distance := make(map[*City]int)
	parent := make(map[*City]*City)

	for _, city := range g.cities {
		distance[city] = math.MaxInt32
		visited[city] = false
	}

	distance[start] = 0

	for i := 0; i < len(g.cities); i++ {
		current := minDistance(distance, visited)
		visited[current] = true

		for _, edge := range g.edges {
			if edge.from == current && !visited[edge.to] {
				newDistance := distance[current] + edge.price
				if newDistance < distance[edge.to] {
					distance[edge.to] = newDistance
					parent[edge.to] = current
				}
			}
		}
	}

	// Build the path
	current := end
	for current != nil {
		path = append([]*City{current}, path...)
		current = parent[current]
	}

	return path, distance[end]
}

// minDistance знаходження міста з найменшою відстанню від стартового міста
func minDistance(distance map[*City]int, visited map[*City]bool) *City {
	min := math.MaxInt32
	var minCity *City

	for city, dist := range distance {
		if !visited[city] && dist < min {
			min = dist
			minCity = city
		}
	}

	return minCity
}

func main() {
	graph := NewGraph()

	// Додавання міст та рейсів
	cityA := graph.AddCity("A")
	cityB := graph.AddCity("B")
	cityC := graph.AddCity("C")

	graph.AddEdge(cityA, cityB, 10, true)
	graph.AddEdge(cityA, cityC, 15, false)
	graph.AddEdge(cityB, cityC, 5, true)

	// Додавання нового міста D
	cityD := graph.AddCity("D")

	// Вивід стану графу
	fmt.Println("Initial Graph State:")
	graph.PrintGraph()

	// Модифікація графу в різних потоках
	go func() {
		graph.ChangeTicketPrice(cityA, cityB, 20)
	}()

	go func() {
		graph.RemoveEdge(cityA, cityC)
		graph.AddEdge(cityB, cityA, 8, false)
	}()

	go func() {
		newCityE := graph.AddCity("E")
		graph.AddEdge(cityC, newCityE, 12, true)
		graph.RemoveCity(cityB)
	}()

	// Затримка, щоб потоки мали час виконатися
	fmt.Println("Waiting for goroutines to finish...")
	fmt.Scanln()

	// Вивід оновленого стану графу
	fmt.Println("\nUpdated Graph State:")
	graph.PrintGraph()

	// Знаходження шляху та ціни між містами
	path, totalPrice := graph.FindPathAndPrice(cityA, cityD)

	// Вивід результату
	fmt.Printf("\nPath from %s to %s: %v\n", cityA.name, cityD.name, path)
	fmt.Printf("Total Price: %d\n", totalPrice)
}

// PrintGraph вивід стану графу
func (g *Graph) PrintGraph() {
	g.lock.RLock()
	defer g.lock.RUnlock()

	fmt.Println("Cities:")
	for _, city := range g.cities {
		fmt.Printf("%s ", city.name)
	}
	fmt.Println("\nEdges:")
	for _, edge := range g.edges {
		fmt.Printf("%s to %s (Price: %d)\n", edge.from.name, edge.to.name, edge.price)
	}
}
