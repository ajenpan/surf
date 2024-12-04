package lobby

import "sync"

var TableStoreInstance *TableStore = NewTableStore()

func NewTableStore() *TableStore {
	return &TableStore{
		nodeHoldTables: make(map[uint32]TableUIDSet),
	}
}

type TableUIDSet = map[TableUIDT]struct{}

type TableStore struct {
	tablesMap sync.Map

	mux            sync.RWMutex
	nodeHoldTables map[uint32]TableUIDSet
}

func (s *TableStore) FindTable(tableuid TableUIDT) *Table {
	if table, ok := s.tablesMap.Load(tableuid); ok {
		return table.(*Table)
	}
	return nil
}

func (s *TableStore) StoreTable(table *Table) bool {
	_, loaded := s.tablesMap.LoadOrStore(table.tuid, table)
	if !loaded {
		return false
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	appid := table.battleNodeId
	if _, ok := s.nodeHoldTables[appid]; !ok {
		s.nodeHoldTables[appid] = make(TableUIDSet)
	}
	s.nodeHoldTables[appid][table.tuid] = struct{}{}
	return true
}

func (s *TableStore) RemoveTable(tableuid TableUIDT) {
	table, loaded := s.tablesMap.LoadAndDelete(tableuid)
	if !loaded {
		return
	}

	pTable := table.(*Table)

	s.mux.Lock()
	defer s.mux.Unlock()

	if _, ok := s.nodeHoldTables[pTable.battleNodeId]; !ok {
		return
	}
	delete(s.nodeHoldTables[pTable.battleNodeId], tableuid)

	pTable.battleNodeId = 0
	pTable.tuid = 0
	pTable.idx = 0
}

func (s *TableStore) RemoveTablesByLoigcAppid(nodeid uint32, onAfterDel func(table *Table)) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, ok := s.nodeHoldTables[nodeid]; !ok {
		return
	}

	tables := s.nodeHoldTables[nodeid]
	delete(s.nodeHoldTables, nodeid)

	for tableuid := range tables {
		table, loaded := s.tablesMap.LoadAndDelete(tableuid)
		if !loaded {
			return
		}

		pTable := table.(*Table)
		if onAfterDel != nil {
			onAfterDel(pTable)
		}

		pTable.battleNodeId = 0
		pTable.tuid = 0
		pTable.idx = 0
	}
}
