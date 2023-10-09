package middleware

// // MiddlewareFunc defines a function to process middleware.
// type MiddlewareFunc func(resmap.ResMap) error

// // ApplyURLs applies all the URLs defined in this module to the provided ResMap.
// func ApplyURLs(s string) MiddlewareFunc {
// 	return func(rm resmap.ResMap) error {
// 		// Create a list of ingress resources to transform
// 		ingressResources := rm.GetMatchingResourcesByAnyId(func(id resid.ResId) bool {
// 			if id.Group == "networking.k8s.io" && id.Kind == "Ingress" {
// 				return true
// 			}
// 			return false
// 		})

// 		// Loop throught the ingresses found, change the host field, and finally apply the patch
// 		for i, ing := range ingressResources {
// 			_, err := ing.Pipe(
// 				kyaml.LookupCreate(kyaml.MappingNode, "spec", "rules", fmt.Sprint(i)),
// 				kyaml.SetField("host", kyaml.NewScalarRNode(s)),
// 			)
// 			if err != nil {
// 				return nil
// 			}
// 			idSet := resource.MakeIdSet(ingressResources)
// 			err = rm.ApplySmPatch(idSet, ing)
// 			if err != nil {
// 				return nil
// 			}
// 		}
// 		return nil
// 	}
// }

// // ApplySecrets applies all secrets defined in this module to the provided ResMap.
// // Searches through the given resmap for Secret resources, updating/adding secrets on this module.
// // The Secret resource to update is determined by the secret key name itself.
// // This function only adds or updates values in the Secret resource if the key matches that of the module.
// func ApplySecrets(secrets map[string]string) MiddlewareFunc {

// 	return func(rm resmap.ResMap) error {
// 		// Create a list of Secret resources to transform
// 		secretResources := rm.GetMatchingResourcesByAnyId(func(id resid.ResId) bool {
// 			return id.Kind == "Secret"
// 		})

// 		// Range over each secret. If the key matches that of a Secret resource then replace it's value with a strategic merge patch
// 		for k, v := range secrets {
// 			for _, secRes := range secretResources {
// 				_, err := secRes.Pipe(
// 					kyaml.Lookup("data", k),
// 					kyaml.Set(kyaml.NewScalarRNode(v)),
// 				)
// 				if err != nil {
// 					return err
// 				}
// 				idSet := resource.MakeIdSet([]*resource.Resource{secRes})
// 				err = rm.ApplySmPatch(idSet, secRes)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		}
// 		return nil
// 	}
// }
