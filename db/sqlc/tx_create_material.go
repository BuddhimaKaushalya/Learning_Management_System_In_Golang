package db

import "context"

type CreateMaterialTxParams struct {
	CreateMaterialParams
	AfterCreate func(material Material) error
}

type CreateMaterialTxResult struct {
	Material Material
}

func (store *SQLStore) CreateMaterialTx(ctx context.Context, arg CreateMaterialTxParams) (CreateMaterialTxResult, error) {
	var result CreateMaterialTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Material, err = q.CreateMaterial(ctx, arg.CreateMaterialParams)
		if err != nil {
			return err
		}

		return arg.AfterCreate(result.Material)
	})

	return result, err
}
