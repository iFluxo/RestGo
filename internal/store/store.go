package store

import (
	"sort"
	"sync"
	"time"
)

// Item is the main resource of this API.
type Item struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ItemInput is the payload accepted from clients on create/update.
type ItemInput struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// Store keeps data in memory and is safe for concurrent use.
type Store struct {
	mu     sync.RWMutex
	items  map[int64]Item
	nextID int64
}

func New() *Store {
	return &Store{
		items:  make(map[int64]Item),
		nextID: 1,
	}
}

// Seed fills the store with sample data so the API can be tried immediately.
func (s *Store) Seed() {
	s.Create(ItemInput{Name: "Milk Coffee", Price: 18000})
	s.Create(ItemInput{Name: "Toast", Price: 15000})
}

func (s *Store) List() []Item {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Item, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items
}

func (s *Store) Get(id int64) (Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[id]
	return item, ok
}

func (s *Store) Create(in ItemInput) Item {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	item := Item{
		ID:        s.nextID,
		Name:      in.Name,
		Price:     in.Price,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.nextID++
	s.items[item.ID] = item
	return item
}

func (s *Store) Update(id int64, in ItemInput) (Item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[id]
	if !ok {
		return Item{}, false
	}

	item.Name = in.Name
	item.Price = in.Price
	item.UpdatedAt = time.Now().UTC()
	s.items[id] = item
	return item, true
}

func (s *Store) Delete(id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.items[id]; !ok {
		return false
	}
	delete(s.items, id)
	return true
}
