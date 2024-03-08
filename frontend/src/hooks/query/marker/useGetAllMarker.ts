import { useQuery } from "@tanstack/react-query";
import getAllMarker from "../../../api/markers/getAllMarker";

const useGetAllMarker = () => {
  return useQuery({
    queryKey: ["marker", "all"],
    queryFn: getAllMarker,
    retry: false,
  });
};

export default useGetAllMarker;
