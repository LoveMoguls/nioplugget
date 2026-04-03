export const load = ({ params }: { params: { id: string } }) => {
	return { studentId: params.id };
};
