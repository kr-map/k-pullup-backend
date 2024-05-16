import approveReport from "@/api/report/approveReport";
import useMapStore from "@/store/useMapStore";
import { useMutation, useQueryClient } from "@tanstack/react-query";
// TODO: 오버레이 위치 변경 같이

const useApproveReport = (markerId: number, lat: number, lng: number) => {
  const queryClient = useQueryClient();
  const { markers, overlay } = useMapStore();

  const filtering = () => {
    if (!markers || !overlay) return;
    const newPosition = new window.kakao.maps.LatLng(lat, lng);
    const marker = markers.find((value) => Number(value.Gb) === markerId);

    if (marker) {
      marker.setPosition(newPosition);
      overlay.setPosition(newPosition);
    }
  };

  return useMutation({
    mutationFn: approveReport,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["marker", "report", "me"] });
      queryClient.invalidateQueries({ queryKey: ["marker", "report", "all"] });
      queryClient.invalidateQueries({
        queryKey: ["marker", "report", "formarker"],
      });
      queryClient.invalidateQueries({
        queryKey: ["marker", "report", "formarker", markerId],
      });
      queryClient.invalidateQueries({
        queryKey: ["marker", markerId],
      });
      filtering();
    },
  });
};

export default useApproveReport;
