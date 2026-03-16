package order

// this is the old cyclic counter
func MergeOrder(id string, local Order, received Order, peersAlive []string) Order {
	merged := local
	merged = resolveUnknown(merged, received)
	merged = guardIllegalRollback(merged, received)
	merged = adoptHigherStatus(merged, received)
	merged = advanceBarrier(merged, id, peersAlive, received)
	return merged
}

func resolveUnknown(merged Order, received Order) Order {
	if merged.Status != OS_Unk {
		return merged
	}
	switch received.Status {
	case OS_Unk:
		return merged
	case OS_None:
		return Order{Status: OS_None, Barrier: []string{}}
	case OS_Unconst, OS_Const:
		return Order{
			Status:  OS_Unconst,
			Barrier: copyBarrier(received.Barrier),
		}
	case OS_Fin:
		return Order{
			Status:  OS_Fin,
			Barrier: copyBarrier(received.Barrier),
		}
	}
	return merged
}

func guardIllegalRollback(merged Order, received Order) Order {
	// Guard: never roll back from OS_None to OS_Fin
	// This handles stale OS_Fin broadcasts arriving after we already cleared
	if merged.Status == OS_None && received.Status == OS_Fin {
		return merged
	}
	// If we are at OS_Fin and receive OS_None, the cycle completed
	// on another elevator — we should also clear
	if merged.Status == OS_Fin && received.Status == OS_None {
		return Order{Status: OS_None, Barrier: []string{}}
	}
	return merged
}

func adoptHigherStatus(merged Order, received Order) Order {
	if received.Status <= merged.Status {
		return merged
	}

	networkAlreadySettled := received.Status == OS_Const
	weAreStillWaiting := merged.Status == OS_Unconst

	if networkAlreadySettled && weAreStillWaiting {
		return Order{Status: OS_Const, Barrier: []string{}}
	}
	return Order{
		Status:  received.Status,
		Barrier: copyBarrier(received.Barrier),
	}
}

func advanceBarrier(merged Order, id string, peersAlive []string, received Order) Order {
	isWaitingStage := merged.Status == OS_Unconst || merged.Status == OS_Fin
	if !isWaitingStage {
		return merged
	}
	merged.Barrier = mergeUnique(merged.Barrier, received.Barrier)
	merged.Barrier = addIfMissing(merged.Barrier, id)
	merged.Barrier = removeOfflinePeers(merged.Barrier, peersAlive)

	if allAliveAcknowledged(merged.Barrier, peersAlive) {
		merged = advanceStatus(merged)
	}
	return merged
}

func advanceStatus(order Order) Order {
	transitions := map[OrderStatus]OrderStatus{
		OS_Unconst: OS_Const,
		OS_Fin:     OS_None,
	}
	next, validTransition := transitions[order.Status]
	if !validTransition {
		return order
	}
	return Order{Status: next, Barrier: []string{}}

}

func allAliveAcknowledged(barrier []string, peersAlive []string) bool {
	for _, peer := range peersAlive {
		if !containsID(barrier, peer) {
			return false
		}
	}
	return true
}

func removeOfflinePeers(barrier []string, peersAlive []string) []string {
	stillOnline := make([]string, 0, len(barrier))
	for _, id := range barrier {
		if containsID(peersAlive, id) {
			stillOnline = append(stillOnline, id)
		}
	}
	return stillOnline
}

func mergeUnique(a, b []string) []string {
	merged := make([]string, len(a))
	copy(merged, a)
	for _, id := range b {
		if !containsID(merged, id) {
			merged = append(merged, id)
		}
	}
	return merged
}

func addIfMissing(slice []string, id string) []string {
	if containsID(slice, id) {
		return slice
	}
	return append(slice, id)
}

func containsID(slice []string, target string) bool {
	for _, v := range slice {
		if v == target {
			return true
		}
	}
	return false
}

func copyBarrier(b []string) []string {
	c := make([]string, len(b))
	copy(c, b)
	return c
}

func PeersAliveInBarrier(barrier []string, peersAlive []string) bool {
	return allAliveAcknowledged(barrier, peersAlive)
}

func Contains(slice []string, target string) bool {
	return containsID(slice, target)
}

func MergeUnique(a, b []string) []string {
	return mergeUnique(a, b)
}

func CopyBarrier(b []string) []string {
	c := make([]string, len(b))
	copy(c, b)
	return c
}
