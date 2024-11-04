package core

// import "context"

// var _ Repository = testRepository{}

// type testRepository struct {
// }

// // Load implements Repository.
// func (t testRepository) Load(ctx context.Context, id AggregateId, aggregate IAggregate) error {
// 	aggregate.Restore(id, testAggState{MyString: "restored"}, Version(1))
// 	return nil
// }

// // Save implements Repository.
// func (t testRepository) Save(ctx context.Context, aggregate IAggregate) error {
// 	return aggregate.Store(func(as AggregateState, ep EventPack, v Version) error {
// 		return nil
// 	})
// }

// func ETLI() {
// 	repo := testRepository{}
// 	agg := &testAgg{}
// 	repo.Load(context.TODO(), "id", agg)
// 	agg.SingleEventCommand("val")
// 	repo.Save(context.TODO(), agg)
// }
