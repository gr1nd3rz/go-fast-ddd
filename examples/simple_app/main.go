package main

func main() {
	// repo := testRepository{}
	agg := testAgg{}
	// repo.Load(context.TODO(), "id", &agg)
	agg.SingleEventCommand("val")
	// agg.Store(func(as core.AggregateState, ep core.EventPack, v core.Version) error {
	// 	return nil
	// })

	// repo.Save(context.TODO(), &agg)
}

// var _ core.Repository = testRepository{}

// type testRepository struct {
// }

// // Load implements Repository.
// func (t testRepository) Load(ctx context.Context, id core.AggregateId, aggregate core.IAggregate) error {
// 	aggregate.Restore(id, testAggState{MyString: "restored"}, core.Version(1))
// 	return nil
// }

// // Save implements Repository.
// func (t testRepository) Save(ctx context.Context, aggregate core.IAggregate) error {
// 	return aggregate.Store(func(as core.AggregateState, ep core.EventPack, v core.Version) error {
// 		return nil
// 	})
// }
